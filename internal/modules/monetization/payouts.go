package monetization

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var payoutSecretVersion = []byte("v1")

func derivePayoutKey(pepper []byte) [32]byte {
	h := sha256.New()
	h.Write([]byte("treeblog/payout-storage/v1"))
	h.Write([]byte{0x00})
	h.Write(pepper)
	var key [32]byte
	copy(key[:], h.Sum(nil))
	return key
}

// encryptAccount seals a bank account number with AES-256-GCM.
func encryptAccount(pepper []byte, account string) ([]byte, error) {
	if len(pepper) < 32 || account == "" {
		return nil, errors.New("pepper and account are required")
	}
	key := derivePayoutKey(pepper)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nil, nonce, []byte(account), nil)
	out := make([]byte, 0, len(payoutSecretVersion)+len(nonce)+len(sealed))
	out = append(out, payoutSecretVersion...)
	out = append(out, nonce...)
	out = append(out, sealed...)
	return out, nil
}

// decryptAccount opens a sealed bank account number.
func decryptAccount(pepper, sealed []byte) (string, error) {
	if len(pepper) < 32 || len(sealed) <= len(payoutSecretVersion) {
		return "", errors.New("invalid sealed account payload")
	}
	for i, b := range payoutSecretVersion {
		if sealed[i] != b {
			return "", errors.New("unsupported account secret version")
		}
	}
	key := derivePayoutKey(pepper)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	rest := sealed[len(payoutSecretVersion):]
	if len(rest) <= gcm.NonceSize() {
		return "", errors.New("invalid sealed account payload")
	}
	nonce, ciphertext := rest[:gcm.NonceSize()], rest[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("account authentication failed")
	}
	return string(plain), nil
}

// Payout is one payout request row (account numbers are never returned).
type Payout struct {
	ID                  string     `json:"id"`
	CreatorUserID       string     `json:"creator_user_id"`
	BlogID              string     `json:"blog_id"`
	PeriodStart         time.Time  `json:"period_start"`
	PeriodEnd           time.Time  `json:"period_end"`
	GrossAmountMinor    int64      `json:"gross_amount_minor"`
	WithholdingTaxMinor int64      `json:"withholding_tax_minor"`
	NetAmountMinor      int64      `json:"net_amount_minor"`
	BankCode            string     `json:"bank_code"`
	AccountHolder       string     `json:"account_holder"`
	Status              string     `json:"status"`
	RequestedAt         time.Time  `json:"requested_at"`
}

// RequestPayoutInput carries the payout request payload.
type RequestPayoutInput struct {
	BlogID        string `json:"blog_id"`
	GrossAmountMinor int64 `json:"gross_amount_minor"`
	BankCode      string `json:"bank_code"`
	BankAccount   string `json:"bank_account"`
	AccountHolder string `json:"account_holder"`
}

// RequestPayout files a creator payout request. Only the blog owner can
// request, and the amount cannot exceed their settled escrow balance.
func (s *Service) RequestPayout(ctx context.Context, creatorID string, input RequestPayoutInput) (*Payout, error) {
	if strings.TrimSpace(input.BlogID) == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_PAYOUT", Message: "Blog id is required"}
	}
	if input.GrossAmountMinor < 10000 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "PAYOUT_TOO_SMALL", Message: "Minimum payout is 10000 KRW"}
	}
	if len(input.BankCode) != 3 || strings.TrimSpace(input.BankAccount) == "" || strings.TrimSpace(input.AccountHolder) == "" {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_BANK_DETAILS", Message: "Bank code (3 digits), account, and holder are required"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var ownerID string
	if err := tx.QueryRow(ctx, `SELECT owner_user_id FROM blogs WHERE id = $1`, input.BlogID).Scan(&ownerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: http.StatusNotFound, Code: "BLOG_NOT_FOUND", Message: "Blog not found"}
		}
		return nil, err
	}
	if ownerID != creatorID {
		return nil, &Fault{Status: http.StatusForbidden, Code: "NOT_BLOG_OWNER", Message: "Only the blog owner can request payouts"}
	}
	var balance int64
	err = tx.QueryRow(ctx, `
		SELECT COALESCE(SUM(CASE WHEN direction = 'CREDIT' THEN amount_minor ELSE -amount_minor END), 0)
		FROM ledger_entries
		WHERE owner_user_id = $1 AND blog_id = $2 AND account_code = 'CREATOR_PENDING_ESCROW'
	`, creatorID, input.BlogID).Scan(&balance)
	if err != nil {
		return nil, err
	}
	if input.GrossAmountMinor > balance {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INSUFFICIENT_ESCROW_BALANCE", Message: "Amount exceeds the settled escrow balance"}
	}
	sealed, err := encryptAccount(s.pepper, strings.TrimSpace(input.BankAccount))
	if err != nil {
		return nil, err
	}
	withholding := input.GrossAmountMinor * WithholdingPermille / 1000
	now := time.Now().UTC()
	var payout Payout
	err = tx.QueryRow(ctx, `
		INSERT INTO payouts (creator_user_id, blog_id, period_start, period_end,
		                     gross_amount_minor, withholding_tax_minor, net_amount_minor,
		                     bank_code, bank_account_enc, account_holder, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'REQUESTED')
		RETURNING id, creator_user_id, blog_id, period_start, period_end,
		          gross_amount_minor, withholding_tax_minor, net_amount_minor,
		          bank_code, account_holder, status, requested_at
	`, creatorID, input.BlogID, now.AddDate(0, -1, 0), now,
		input.GrossAmountMinor, withholding, input.GrossAmountMinor-withholding,
		input.BankCode, sealed, strings.TrimSpace(input.AccountHolder)).Scan(
		&payout.ID, &payout.CreatorUserID, &payout.BlogID, &payout.PeriodStart, &payout.PeriodEnd,
		&payout.GrossAmountMinor, &payout.WithholdingTaxMinor, &payout.NetAmountMinor,
		&payout.BankCode, &payout.AccountHolder, &payout.Status, &payout.RequestedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &payout, nil
}

// ListPayouts returns payouts for finance review with offset paging.
func (s *Service) ListPayouts(ctx context.Context, status string, limit, offset int) ([]Payout, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}
	query := `
		SELECT id, creator_user_id, blog_id, period_start, period_end,
		       gross_amount_minor, withholding_tax_minor, net_amount_minor,
		       bank_code, account_holder, status, requested_at
		FROM payouts
	`
	args := []any{}
	if status != "" {
		query += ` WHERE status = $1`
		args = append(args, strings.ToUpper(status))
		query += ` ORDER BY requested_at DESC LIMIT $2 OFFSET $3`
		args = append(args, limit, offset)
	} else {
		query += ` ORDER BY requested_at DESC LIMIT $1 OFFSET $2`
		args = append(args, limit, offset)
	}
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]Payout, 0)
	for rows.Next() {
		var payout Payout
		if err := rows.Scan(&payout.ID, &payout.CreatorUserID, &payout.BlogID, &payout.PeriodStart,
			&payout.PeriodEnd, &payout.GrossAmountMinor, &payout.WithholdingTaxMinor, &payout.NetAmountMinor,
			&payout.BankCode, &payout.AccountHolder, &payout.Status, &payout.RequestedAt); err != nil {
			return nil, err
		}
		items = append(items, payout)
	}
	return items, rows.Err()
}

// DecidePayout approves, holds, or rejects a payout request.
func (s *Service) DecidePayout(ctx context.Context, payoutID, decision, reason string) (*Payout, error) {
	decision = strings.ToUpper(strings.TrimSpace(decision))
	var next string
	switch decision {
	case "APPROVE":
		next = "APPROVED"
	case "HOLD":
		next = "ON_HOLD"
	case "REJECT":
		next = "REJECTED"
	default:
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_DECISION", Message: "Decision must be APPROVE, HOLD, or REJECT"}
	}
	if len(reason) > 500 {
		return nil, &Fault{Status: http.StatusBadRequest, Code: "INVALID_REASON", Message: "Reason exceeds 500 characters"}
	}
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	var payout Payout
	err := s.pool.QueryRow(ctx, `
		UPDATE payouts SET status = $2, hold_reason = NULLIF($3, ''), decided_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status IN ('REQUESTED', 'ON_HOLD')
		RETURNING id, creator_user_id, blog_id, period_start, period_end,
		          gross_amount_minor, withholding_tax_minor, net_amount_minor,
		          bank_code, account_holder, status, requested_at
	`, payoutID, next, reason).Scan(
		&payout.ID, &payout.CreatorUserID, &payout.BlogID, &payout.PeriodStart, &payout.PeriodEnd,
		&payout.GrossAmountMinor, &payout.WithholdingTaxMinor, &payout.NetAmountMinor,
		&payout.BankCode, &payout.AccountHolder, &payout.Status, &payout.RequestedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusConflict, Code: "PAYOUT_NOT_DECIDABLE", Message: "Payout is not in a decidable state"}
	}
	return &payout, err
}

// RevealPayoutAccount decrypts the bank account for a payout. Callers must
// gate this behind finance roles and audit access.
func (s *Service) RevealPayoutAccount(ctx context.Context, payoutID string) (bankCode, account, holder string, err error) {
	if err := s.requirePool(); err != nil {
		return "", "", "", err
	}
	var sealed []byte
	err = s.pool.QueryRow(ctx, `SELECT bank_code, bank_account_enc, account_holder
		FROM payouts WHERE id = $1`, payoutID).Scan(&bankCode, &sealed, &holder)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", "", &Fault{Status: http.StatusNotFound, Code: "PAYOUT_NOT_FOUND", Message: "Payout not found"}
	}
	if err != nil {
		return "", "", "", err
	}
	account, err = decryptAccount(s.pepper, sealed)
	if err != nil {
		return "", "", "", &Fault{Status: http.StatusInternalServerError, Code: "ACCOUNT_UNAVAILABLE", Message: "Bank account cannot be decrypted"}
	}
	return bankCode, account, holder, nil
}

// MarkPayoutPaid records a completed bank transfer and posts settlement
// ledger entries (escrow out, disbursement + withholding out).
func (s *Service) MarkPayoutPaid(ctx context.Context, payoutID string) (*Payout, error) {
	if err := s.requirePool(); err != nil {
		return nil, err
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var payout Payout
	err = tx.QueryRow(ctx, `
		SELECT id, creator_user_id, blog_id, period_start, period_end,
		       gross_amount_minor, withholding_tax_minor, net_amount_minor,
		       bank_code, account_holder, status, requested_at
		FROM payouts WHERE id = $1 FOR UPDATE`, payoutID).Scan(
		&payout.ID, &payout.CreatorUserID, &payout.BlogID, &payout.PeriodStart, &payout.PeriodEnd,
		&payout.GrossAmountMinor, &payout.WithholdingTaxMinor, &payout.NetAmountMinor,
		&payout.BankCode, &payout.AccountHolder, &payout.Status, &payout.RequestedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, &Fault{Status: http.StatusNotFound, Code: "PAYOUT_NOT_FOUND", Message: "Payout not found"}
	}
	if err != nil {
		return nil, err
	}
	if payout.Status != "APPROVED" {
		return nil, &Fault{Status: http.StatusConflict, Code: "PAYOUT_NOT_APPROVED", Message: "Only approved payouts can be marked paid"}
	}
	entries := []struct {
		account, direction string
		amount             int64
	}{
		{"CREATOR_PENDING_ESCROW", "DEBIT", payout.GrossAmountMinor},
		{"TAX_WITHHOLDING_PAYABLE", "CREDIT", payout.WithholdingTaxMinor},
		{"PAYOUT_DISBURSED", "CREDIT", payout.NetAmountMinor},
	}
	for _, entry := range entries {
		if entry.amount <= 0 {
			continue
		}
		if _, err := tx.Exec(ctx, `
			INSERT INTO ledger_entries (owner_user_id, blog_id, account_code, direction, amount_minor, currency, reference_type, reference_id, idempotency_key)
			VALUES ($1, $2, $3, $4, $5, 'KRW', 'PAYOUT_SETTLEMENT', $6, $7)
			ON CONFLICT (idempotency_key) DO NOTHING
		`, payout.CreatorUserID, payout.BlogID, entry.account, entry.direction, entry.amount, payout.ID,
			"ledger-payout-"+payout.ID+"-"+entry.account); err != nil {
			return nil, err
		}
	}
	if _, err := tx.Exec(ctx, `UPDATE payouts SET status = 'PAID', updated_at = NOW() WHERE id = $1`, payoutID); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	payout.Status = "PAID"
	return &payout, nil
}
