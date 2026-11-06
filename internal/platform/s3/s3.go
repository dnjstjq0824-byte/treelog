// Package s3 is a minimal S3-compatible client (AWS Signature Version 4)
// using only the standard library. It uses path-style URLs so the same code
// works against AWS S3 and MinIO.
package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Config describes the S3-compatible endpoint and credentials.
type Config struct {
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
}

// Client signs and executes S3 requests.
type Client struct {
	cfg      Config
	base     *url.URL
	http     *http.Client
	now      func() time.Time
}

// New creates a Client. Credentials are required: anonymous object storage
// would silently turn uploads into dead links.
func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.Region == "" || cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, errors.New("s3 endpoint, region, and credentials are required")
	}
	base, err := url.Parse(cfg.Endpoint)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return nil, fmt.Errorf("invalid s3 endpoint %q", cfg.Endpoint)
	}
	return &Client{
		cfg:  cfg,
		base: base,
		http: &http.Client{Timeout: 60 * time.Second},
		now:  time.Now,
	}, nil
}

// PresignPUT returns a presigned URL the browser uses to upload one object
// directly. The caller must PUT exactly contentType; content length is
// verified server-side on completion via HEAD.
func (c *Client) PresignPUT(bucket, key, contentType string, expires time.Duration) (string, error) {
	return c.presign(http.MethodPut, bucket, key, expires, "UNSIGNED-PAYLOAD")
}

// PresignGET returns a presigned URL for downloading one object.
func (c *Client) PresignGET(bucket, key string, expires time.Duration) (string, error) {
	return c.presign(http.MethodGet, bucket, key, expires, "UNSIGNED-PAYLOAD")
}

func (c *Client) presign(method, bucket, key string, expires time.Duration, payloadHash string) (string, error) {
	if bucket == "" || key == "" {
		return "", errors.New("s3 bucket and key are required")
	}
	if expires <= 0 || expires > 7*24*time.Hour {
		return "", errors.New("s3 presign expiry must be between 1s and 7 days")
	}
	now := c.now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, c.cfg.Region)

	canonicalURI := "/" + bucket + "/" + encodeKey(key)
	query := url.Values{}
	query.Set("X-Amz-Algorithm", "AWS4-HMAC-SHA256")
	query.Set("X-Amz-Credential", c.cfg.AccessKey+"/"+credentialScope)
	query.Set("X-Amz-Date", amzDate)
	query.Set("X-Amz-Expires", strconv.Itoa(int(expires.Seconds())))
	query.Set("X-Amz-SignedHeaders", "host")
	canonicalQuery := canonicalQueryString(query)

	canonicalHeaders := "host:" + c.base.Host + "\n"
	canonicalRequest := strings.Join([]string{
		method, canonicalURI, canonicalQuery, canonicalHeaders, "host", payloadHash,
	}, "\n")

	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256", amzDate, credentialScope, sha256Hex(canonicalRequest),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(c.signingKey(dateStamp), stringToSign))

	signed := c.objectURL(bucket, key)
	signed.RawQuery = canonicalQuery + "&X-Amz-Signature=" + signature
	return signed.String(), nil
}

// objectURL builds the path-style object URL, keeping the SigV4-encoded
// form in RawPath so url.String never double-encodes it.
func (c *Client) objectURL(bucket, key string) url.URL {
	target := *c.base
	target.Path = "/" + bucket + "/" + strings.TrimPrefix(key, "/")
	target.RawPath = "/" + bucket + "/" + encodeKey(key)
	return target
}

// Head returns the object size in bytes.
func (c *Client) Head(ctx context.Context, bucket, key string) (int64, error) {
	resp, err := c.do(ctx, http.MethodHead, bucket, key, "", nil, nil)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("s3 head %s/%s: status %d", bucket, key, resp.StatusCode)
	}
	size, err := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	if err != nil || size < 0 {
		return 0, fmt.Errorf("s3 head %s/%s: bad content length", bucket, key)
	}
	return size, nil
}

// Get downloads at most maxBytes+1 bytes; callers treat over-cap responses
// as an error without trusting the full payload.
func (c *Client) Get(ctx context.Context, bucket, key string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("s3 get requires a positive byte cap")
	}
	resp, err := c.do(ctx, http.MethodGet, bucket, key, "", nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("s3 get %s/%s: status %d", bucket, key, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("s3 get %s/%s: %w", bucket, key, err)
	}
	if int64(len(body)) > maxBytes {
		return nil, fmt.Errorf("s3 get %s/%s: object exceeds %d bytes", bucket, key, maxBytes)
	}
	return body, nil
}

// GetRange downloads one byte range (offset, length) of an object.
func (c *Client) GetRange(ctx context.Context, bucket, key string, offset, length int64) ([]byte, error) {
	if offset < 0 || length <= 0 {
		return nil, errors.New("s3 range requires a non-negative offset and positive length")
	}
	headers := map[string]string{
		"Range": fmt.Sprintf("bytes=%d-%d", offset, offset+length-1),
	}
	resp, err := c.do(ctx, http.MethodGet, bucket, key, "", nil, headers)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("s3 range %s/%s: status %d", bucket, key, resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, length+1))
	if err != nil {
		return nil, fmt.Errorf("s3 range %s/%s: %w", bucket, key, err)
	}
	return body, nil
}

// Put uploads (or overwrites) one object.
func (c *Client) Put(ctx context.Context, bucket, key, contentType string, body []byte) error {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	resp, err := c.do(ctx, http.MethodPut, bucket, key, contentType, body, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated &&
		resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("s3 put %s/%s: status %d", bucket, key, resp.StatusCode)
	}
	return nil
}

func (c *Client) do(ctx context.Context, method, bucket, key, contentType string, body []byte, extra map[string]string) (*http.Response, error) {
	if bucket == "" || key == "" {
		return nil, errors.New("s3 bucket and key are required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	target := c.objectURL(bucket, key)
	target.RawQuery = ""
	canonicalURI := "/" + bucket + "/" + encodeKey(key)

	var reader io.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}
	req, err := http.NewRequestWithContext(ctx, method, target.String(), reader)
	if err != nil {
		return nil, err
	}
	payloadHash := sha256Hex(string(body))
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	now := c.now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")
	req.Header.Set("X-Amz-Date", amzDate)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	for name, value := range extra {
		req.Header.Set(name, value)
	}
	req.Header.Set("Authorization", c.headerAuth(method, canonicalURI, payloadHash, amzDate, dateStamp))
	return c.http.Do(req)
}

func (c *Client) headerAuth(method, canonicalURI, payloadHash, amzDate, dateStamp string) string {
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, c.cfg.Region)
	signedHeaders := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	canonicalHeaders := "host:" + c.base.Host + "\n" +
		"x-amz-content-sha256:" + payloadHash + "\n" +
		"x-amz-date:" + amzDate + "\n"
	canonicalRequest := strings.Join([]string{
		method, canonicalURI, "", canonicalHeaders,
		strings.Join(signedHeaders, ";"), payloadHash,
	}, "\n")
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256", amzDate, credentialScope, sha256Hex(canonicalRequest),
	}, "\n")
	signature := hex.EncodeToString(hmacSHA256(c.signingKey(dateStamp), stringToSign))
	return fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.cfg.AccessKey, credentialScope, strings.Join(signedHeaders, ";"), signature)
}

func (c *Client) signingKey(dateStamp string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+c.cfg.SecretKey), dateStamp)
	kRegion := hmacSHA256(kDate, c.cfg.Region)
	kService := hmacSHA256(kRegion, "s3")
	return hmacSHA256(kService, "aws4_request")
}

// encodeKey percent-encodes every path segment except RFC 3986 unreserved
// characters, using uppercase hex as SigV4 requires.
func encodeKey(key string) string {
	segments := strings.Split(strings.TrimPrefix(key, "/"), "/")
	for i, segment := range segments {
		segments[i] = encodeSegment(segment)
	}
	return strings.Join(segments, "/")
}

func encodeSegment(segment string) string {
	var sb strings.Builder
	for i := 0; i < len(segment); i++ {
		b := segment[i]
		if (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') ||
			(b >= '0' && b <= '9') || b == '-' || b == '_' || b == '.' || b == '~' {
			sb.WriteByte(b)
		} else {
			fmt.Fprintf(&sb, "%%%02X", b)
		}
	}
	return sb.String()
}

func canonicalQueryString(values url.Values) string {
	// url.Values.Encode sorts keys; SigV4 additionally requires %20 for spaces.
	return strings.ReplaceAll(values.Encode(), "+", "%20")
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}

func sha256Hex(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}
