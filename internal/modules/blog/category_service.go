package blog

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type BlogCategory struct {
	ID          string         `json:"id"`
	BlogID      string         `json:"blog_id"`
	ParentID    *string        `json:"parent_id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	Description string         `json:"description"`
	TopicCode   *string        `json:"topic_code"`
	ListStyle   string         `json:"list_style"`
	Visibility  string         `json:"visibility"`
	SortOrder   int            `json:"sort_order"`
	PostCount   int            `json:"post_count"`
	Children    []BlogCategory `json:"children,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type CreateCategoryInput struct {
	ParentID    *string `json:"parent_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description string  `json:"description"`
	TopicCode   *string `json:"topic_code"`
	ListStyle   string  `json:"list_style"`
	Visibility  string  `json:"visibility"`
	SortOrder   int     `json:"sort_order"`
}

type ReorderCategoryItem struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parent_id"`
	SortOrder int     `json:"sort_order"`
}

func (s *Service) ListCategories(ctx context.Context, blogID string) ([]BlogCategory, error) {
	rows, err := s.db.Query(ctx, `SELECT id, blog_id, parent_id, name, slug, description,
		topic_code, list_style, visibility, sort_order, post_count, created_at, updated_at
		FROM blog_categories
		WHERE blog_id = $1
		ORDER BY sort_order ASC, created_at ASC`, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var all []BlogCategory
	for rows.Next() {
		var c BlogCategory
		if err := rows.Scan(
			&c.ID, &c.BlogID, &c.ParentID, &c.Name, &c.Slug, &c.Description,
			&c.TopicCode, &c.ListStyle, &c.Visibility, &c.SortOrder, &c.PostCount, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		all = append(all, c)
	}

	// Build 2-depth tree: roots and their children
	roots := make([]BlogCategory, 0)
	childrenMap := make(map[string][]BlogCategory)

	for _, item := range all {
		if item.ParentID != nil && *item.ParentID != "" {
			childrenMap[*item.ParentID] = append(childrenMap[*item.ParentID], item)
		}
	}
	for _, item := range all {
		if item.ParentID == nil || *item.ParentID == "" {
			item.Children = childrenMap[item.ID]
			roots = append(roots, item)
		}
	}
	return roots, nil
}

func (s *Service) CreateCategory(ctx context.Context, blogID string, input CreateCategoryInput) (*BlogCategory, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 1 || len(name) > 80 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Category name must be between 1 and 80 characters"}
	}
	slug := strings.ToLower(strings.TrimSpace(input.Slug))
	if len(slug) < 1 || len(slug) > 100 {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Category slug must be between 1 and 100 characters"}
	}

	listStyle := strings.ToUpper(strings.TrimSpace(input.ListStyle))
	if listStyle == "" {
		listStyle = "CARD"
	}
	if listStyle != "CARD" && listStyle != "LIST" && listStyle != "GALLERY" && listStyle != "COMPACT" {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid list style"}
	}

	visibility := strings.ToUpper(strings.TrimSpace(input.Visibility))
	if visibility == "" {
		visibility = "PUBLIC"
	}
	if visibility != "PUBLIC" && visibility != "NEIGHBOR_ONLY" && visibility != "MUTUAL_ONLY" && visibility != "PRIVATE" {
		return nil, &Fault{Status: 400, Code: "INVALID_REQUEST_PAYLOAD", Message: "Invalid visibility"}
	}

	// 2-depth hierarchy check: if parent_id is supplied, the parent MUST be a root category (parent_id IS NULL)
	if input.ParentID != nil && *input.ParentID != "" {
		var parentParentID *string
		var parentBlogID string
		err := s.db.QueryRow(ctx, "SELECT blog_id, parent_id FROM blog_categories WHERE id=$1", *input.ParentID).Scan(&parentBlogID, &parentParentID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Parent category not found"}
		}
		if err != nil {
			return nil, err
		}
		if parentBlogID != blogID {
			return nil, &Fault{Status: 403, Code: "INSUFFICIENT_BLOG_ROLE", Message: "Parent category belongs to another blog"}
		}
		if parentParentID != nil {
			return nil, &Fault{Status: 400, Code: "MAX_CATEGORY_DEPTH_EXCEEDED", Message: "Cannot create category deeper than 2 levels"}
		}
	}

	var c BlogCategory
	c.ID = uuid.NewString()
	c.BlogID = blogID
	c.ParentID = input.ParentID
	c.Name = name
	c.Slug = slug
	c.Description = strings.TrimSpace(input.Description)
	c.TopicCode = input.TopicCode
	c.ListStyle = listStyle
	c.Visibility = visibility
	c.SortOrder = input.SortOrder

	err := s.db.QueryRow(ctx, `INSERT INTO blog_categories (
		id, blog_id, parent_id, name, slug, description, topic_code, list_style, visibility, sort_order
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	RETURNING created_at, updated_at`,
		c.ID, blogID, c.ParentID, c.Name, c.Slug, c.Description, c.TopicCode, c.ListStyle, c.Visibility, c.SortOrder,
	).Scan(&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, &Fault{Status: 409, Code: "DUPLICATE_CATEGORY_SLUG", Message: "Category with this slug already exists in blog"}
		}
		return nil, err
	}

	return &c, nil
}

func (s *Service) ReorderCategories(ctx context.Context, blogID string, items []ReorderCategoryItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		// Prevent cycle (cannot set parent to self)
		if item.ParentID != nil && *item.ParentID == item.ID {
			return &Fault{Status: 400, Code: "INVALID_CATEGORY_HIERARCHY", Message: "Category cannot be its own parent"}
		}
		if _, err := tx.Exec(ctx, `UPDATE blog_categories
			SET sort_order = $3, parent_id = $4, updated_at = NOW()
			WHERE id = $1 AND blog_id = $2`,
			item.ID, blogID, item.SortOrder, item.ParentID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *Service) DeleteCategory(ctx context.Context, blogID, categoryID string, targetCategoryID *string) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var slug string
	err = tx.QueryRow(ctx, "SELECT slug FROM blog_categories WHERE id=$1 AND blog_id=$2", categoryID, blogID).Scan(&slug)
	if errors.Is(err, pgx.ErrNoRows) {
		return &Fault{Status: 404, Code: "RESOURCE_NOT_FOUND", Message: "Category not found"}
	}
	if err != nil {
		return err
	}
	if slug == "uncategorized" {
		return &Fault{Status: 400, Code: "CANNOT_DELETE_DEFAULT_CATEGORY", Message: "Default uncategorized category cannot be deleted"}
	}

	// Resolve destination category for posts
	destinationID := ""
	if targetCategoryID != nil && *targetCategoryID != "" && *targetCategoryID != categoryID {
		destinationID = *targetCategoryID
	} else {
		// Find uncategorized category of this blog
		err = tx.QueryRow(ctx, "SELECT id FROM blog_categories WHERE blog_id=$1 AND slug='uncategorized'", blogID).Scan(&destinationID)
		if err != nil {
			return fmt.Errorf("find default category: %w", err)
		}
	}

	// Move posts from deleting category (and any subcategories) to destination
	if _, err := tx.Exec(ctx, `UPDATE posts SET category_id=$1 WHERE category_id=$2`, destinationID, categoryID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE posts SET category_id=$1
		WHERE category_id IN (SELECT id FROM blog_categories WHERE parent_id=$2)`, destinationID, categoryID); err != nil {
		return err
	}

	// Delete subcategories and category
	if _, err := tx.Exec(ctx, "DELETE FROM blog_categories WHERE parent_id=$1 AND blog_id=$2", categoryID, blogID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, "DELETE FROM blog_categories WHERE id=$1 AND blog_id=$2", categoryID, blogID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

