package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/database"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
	"github.com/google/uuid"
)

// PostRepository defines methods to interact with post data
type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id string) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error)
	GetFeed(ctx context.Context, userIDs []string, limit, offset int) ([]*model.Post, error)
}

type postRepository struct {
	db *database.DB
}

// NewPostRepository creates a new PostRepository
func NewPostRepository(db *database.DB) PostRepository {
	return &postRepository{
		db: db,
	}
}

// Create adds a new post
func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	// Generate ID if not provided
	if post.ID == "" {
		post.ID = uuid.New().String()
	}

	query := `
		INSERT INTO post (
			id, user_id, content, embed_urls, hashtags, 
			visibility, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $7
		)
	`

	now := time.Now().UTC()
	post.CreatedAt = now

	// Convert string arrays to PostgreSQL arrays
	embedURLsJSON, err := json.Marshal(post.EmbedURLs)
	if err != nil {
		return fmt.Errorf("marshal embed URLs: %w", err)
	}

	hashtagsJSON, err := json.Marshal(post.Hashtags)
	if err != nil {
		return fmt.Errorf("marshal hashtags: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		post.ID,
		post.UserID,
		post.Content,
		embedURLsJSON,
		hashtagsJSON,
		post.Visibility,
		now,
	)

	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	return nil
}

// GetByID fetches a post by ID
func (r *postRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	query := `
		SELECT 
			p.id, p.user_id, p.content, p.embed_urls, p.hashtags, 
			p.visibility, p.created_at, p.updated_at,
			u.username, u.image_url
		FROM 
			post p
		JOIN 
			"user" u ON p.user_id = u.id
		WHERE 
			p.id = $1
	`

	var post model.Post
	var embedURLsJSON, hashtagsJSON []byte
	var updatedAt sql.NullTime
	var visibilityStr string
	var username string
	var imageURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&embedURLsJSON,
		&hashtagsJSON,
		&visibilityStr,
		&post.CreatedAt,
		&updatedAt,
		&username,
		&imageURL,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound("post", id)
		}
		return nil, fmt.Errorf("query post: %w", err)
	}

	// Parse JSON arrays
	if len(embedURLsJSON) > 0 {
		if err := json.Unmarshal(embedURLsJSON, &post.EmbedURLs); err != nil {
			return nil, fmt.Errorf("unmarshal embed URLs: %w", err)
		}
	}

	if len(hashtagsJSON) > 0 {
		if err := json.Unmarshal(hashtagsJSON, &post.Hashtags); err != nil {
			return nil, fmt.Errorf("unmarshal hashtags: %w", err)
		}
	}

	if updatedAt.Valid {
		post.UpdatedAt = &updatedAt.Time
	}

	post.Visibility = model.Visibility(visibilityStr)

	// Set up basic author information
	author := &model.User{
		ID:       post.UserID,
		Username: username,
	}

	if imageURL.Valid {
		author.ImageURL = &imageURL.String
	}

	post.Author = author

	return &post, nil
}

// Update updates a post
func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	query := `
		UPDATE post
		SET 
			content = $1, 
			embed_urls = $2, 
			hashtags = $3, 
			updated_at = $4
		WHERE 
			id = $5
		RETURNING id
	`

	now := time.Now().UTC()
	postUpdate := now
	post.UpdatedAt = &postUpdate

	// Convert arrays to JSON
	embedURLsJSON, err := json.Marshal(post.EmbedURLs)
	if err != nil {
		return fmt.Errorf("marshal embed URLs: %w", err)
	}

	hashtagsJSON, err := json.Marshal(post.Hashtags)
	if err != nil {
		return fmt.Errorf("marshal hashtags: %w", err)
	}

	var id string
	err = r.db.QueryRowContext(ctx, query,
		post.Content,
		embedURLsJSON,
		hashtagsJSON,
		now,
		post.ID,
	).Scan(&id)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("post", post.ID)
		}
		return fmt.Errorf("update post: %w", err)
	}

	return nil
}

// Delete removes a post
func (r *postRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM post WHERE id = $1 RETURNING id`

	var deletedID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&deletedID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound("post", id)
		}
		return fmt.Errorf("delete post: %w", err)
	}

	return nil
}

// GetByUserID fetches posts by user ID
func (r *postRepository) GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			p.id, p.user_id, p.content, p.embed_urls, p.hashtags, 
			p.visibility, p.created_at, p.updated_at,
			u.username, u.image_url
		FROM 
			post p
		JOIN 
			"user" u ON p.user_id = u.id
		WHERE 
			p.user_id = $1
		ORDER BY 
			p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query user posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post

	for rows.Next() {
		var post model.Post
		var embedURLsJSON, hashtagsJSON []byte
		var updatedAt sql.NullTime
		var visibilityStr string
		var username string
		var imageURL sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&embedURLsJSON,
			&hashtagsJSON,
			&visibilityStr,
			&post.CreatedAt,
			&updatedAt,
			&username,
			&imageURL,
		)

		if err != nil {
			return nil, fmt.Errorf("scan post row: %w", err)
		}

		// Parse JSON arrays
		if len(embedURLsJSON) > 0 {
			if err := json.Unmarshal(embedURLsJSON, &post.EmbedURLs); err != nil {
				return nil, fmt.Errorf("unmarshal embed URLs: %w", err)
			}
		}

		if len(hashtagsJSON) > 0 {
			if err := json.Unmarshal(hashtagsJSON, &post.Hashtags); err != nil {
				return nil, fmt.Errorf("unmarshal hashtags: %w", err)
			}
		}

		if updatedAt.Valid {
			post.UpdatedAt = &updatedAt.Time
		}

		post.Visibility = model.Visibility(visibilityStr)

		// Set up basic author information
		author := &model.User{
			ID:       post.UserID,
			Username: username,
		}

		if imageURL.Valid {
			author.ImageURL = &imageURL.String
		}

		post.Author = author

		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return posts, nil
}

// GetFeed fetches posts for a feed
func (r *postRepository) GetFeed(ctx context.Context, userIDs []string, limit, offset int) ([]*model.Post, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	// If no userIDs provided, return empty slice
	if len(userIDs) == 0 {
		return []*model.Post{}, nil
	}

	query := `
		SELECT 
			p.id, p.user_id, p.content, p.embed_urls, p.hashtags, 
			p.visibility, p.created_at, p.updated_at,
			u.username, u.image_url
		FROM 
			post p
		JOIN 
			"user" u ON p.user_id = u.id
		WHERE 
			p.user_id = ANY($1) AND
			p.visibility = 'public'
		ORDER BY 
			p.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userIDs, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query feed posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post

	for rows.Next() {
		var post model.Post
		var embedURLsJSON, hashtagsJSON []byte
		var updatedAt sql.NullTime
		var visibilityStr string
		var username string
		var imageURL sql.NullString

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&embedURLsJSON,
			&hashtagsJSON,
			&visibilityStr,
			&post.CreatedAt,
			&updatedAt,
			&username,
			&imageURL,
		)

		if err != nil {
			return nil, fmt.Errorf("scan feed post row: %w", err)
		}

		// Parse JSON arrays
		if len(embedURLsJSON) > 0 {
			if err := json.Unmarshal(embedURLsJSON, &post.EmbedURLs); err != nil {
				return nil, fmt.Errorf("unmarshal embed URLs: %w", err)
			}
		}

		if len(hashtagsJSON) > 0 {
			if err := json.Unmarshal(hashtagsJSON, &post.Hashtags); err != nil {
				return nil, fmt.Errorf("unmarshal hashtags: %w", err)
			}
		}

		if updatedAt.Valid {
			post.UpdatedAt = &updatedAt.Time
		}

		post.Visibility = model.Visibility(visibilityStr)

		// Set up basic author information
		author := &model.User{
			ID:       post.UserID,
			Username: username,
		}

		if imageURL.Valid {
			author.ImageURL = &imageURL.String
		}

		post.Author = author

		posts = append(posts, &post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return posts, nil
}
