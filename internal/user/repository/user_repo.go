package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/PeterM45/perfolio-api/internal/common/model"
	"github.com/PeterM45/perfolio-api/internal/platform/database"
	"github.com/PeterM45/perfolio-api/pkg/apperrors"
)

// UserRepository defines methods to interact with user data
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	Update(ctx context.Context, id string, updates map[string]interface{}) error
	Search(ctx context.Context, query string, limit int) ([]*model.User, error)

	AddFollow(ctx context.Context, followerID, followingID string) error
	RemoveFollow(ctx context.Context, followerID, followingID string) error
	IsFollowing(ctx context.Context, followerID, followingID string) (bool, error)
	GetFollowerCount(ctx context.Context, userID string) (int, error)
	GetFollowingCount(ctx context.Context, userID string) (int, error)
	GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*model.User, error)
	GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*model.User, error)
}

type userRepository struct {
	db *database.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *database.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// GetByID fetches a user by ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT 
			id, email, username, first_name, last_name, bio, 
			auth_provider, image_url, is_active, created_at, updated_at
		FROM 
			"users"
		WHERE 
			id = $1
	`

	var user model.User
	var firstName, lastName, bio, imageURL sql.NullString
	var updatedAt sql.NullTime
	var authProviderStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&firstName,
		&lastName,
		&bio,
		&authProviderStr,
		&imageURL,
		&user.IsActive,
		&user.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("user: %s", id))
		}
		return nil, fmt.Errorf("query user: %w", err)
	}

	// Handle null fields
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if bio.Valid {
		user.Bio = &bio.String
	}
	if imageURL.Valid {
		user.ImageURL = &imageURL.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	user.AuthProvider = model.AuthProvider(authProviderStr)

	return &user, nil
}

// GetByUsername fetches a user by username
func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
		SELECT 
			id, email, username, first_name, last_name, bio, 
			auth_provider, image_url, is_active, created_at, updated_at
		FROM 
			"users"
		WHERE 
			username = $1
	`

	var user model.User
	var firstName, lastName, bio, imageURL sql.NullString
	var updatedAt sql.NullTime
	var authProviderStr string

	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&firstName,
		&lastName,
		&bio,
		&authProviderStr,
		&imageURL,
		&user.IsActive,
		&user.CreatedAt,
		&updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("user by username: %s", username))
		}
		return nil, fmt.Errorf("query user by username: %w", err)
	}

	// Handle null fields
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if bio.Valid {
		user.Bio = &bio.String
	}
	if imageURL.Valid {
		user.ImageURL = &imageURL.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	user.AuthProvider = model.AuthProvider(authProviderStr)

	return &user, nil
}

// GetByEmail fetches a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, username, first_name, last_name, bio, auth_provider, 
	password_hash, image_url, is_active, created_at, updated_at 
	FROM users WHERE email = $1`

	var user model.User
	var firstName, lastName, bio, imageURL sql.NullString
	var updatedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Username, &firstName, &lastName, &bio,
		&user.AuthProvider, &user.PasswordHash, &imageURL, &user.IsActive,
		&user.CreatedAt, &updatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.NotFound(fmt.Sprintf("user by email: %s", email))
		}
		return nil, fmt.Errorf("query user by email: %w", err)
	}

	// Handle null fields
	if firstName.Valid {
		user.FirstName = &firstName.String
	}
	if lastName.Valid {
		user.LastName = &lastName.String
	}
	if bio.Valid {
		user.Bio = &bio.String
	}
	if imageURL.Valid {
		user.ImageURL = &imageURL.String
	}
	if updatedAt.Valid {
		user.UpdatedAt = &updatedAt.Time
	}

	//user.AuthProvider = model.AuthProvider(authProviderStr)

	return &user, nil
}

// Create adds a new user
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
        INSERT INTO "users" (
            id, email, username, first_name, last_name, bio,
            auth_provider, password_hash, image_url, is_active, created_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
        )
    `

	now := time.Now().UTC()

	// Handle null fields
	var email, firstName, lastName, bio, imageURL sql.NullString

	if user.Email != "" {
		email = sql.NullString{String: user.Email, Valid: true}
	}
	if user.FirstName != nil {
		firstName = sql.NullString{String: *user.FirstName, Valid: true}
	}
	if user.LastName != nil {
		lastName = sql.NullString{String: *user.LastName, Valid: true}
	}
	if user.Bio != nil {
		bio = sql.NullString{String: *user.Bio, Valid: true}
	}
	if user.ImageURL != nil {
		imageURL = sql.NullString{String: *user.ImageURL, Valid: true}
	}

	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		email,
		user.Username,
		firstName,
		lastName,
		bio,
		string(user.AuthProvider),
		user.PasswordHash, // Add password hash
		imageURL,
		user.IsActive,
		now,
	)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	user.CreatedAt = now

	return nil
}

// Update updates user properties
func (r *userRepository) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	// Start building query
	query := `UPDATE "users" SET updated_at = NOW()`
	params := []interface{}{id} // First param is ID for the WHERE clause
	paramCount := 1

	// Add each field to update
	for key, value := range updates {
		var dbField string

		// Map to database column names
		switch key {
		case "username":
			dbField = "username"
		case "firstName":
			dbField = "first_name"
		case "lastName":
			dbField = "last_name"
		case "bio":
			dbField = "bio"
		case "imageUrl":
			dbField = "image_url"
		case "isActive":
			dbField = "is_active"
		default:
			continue // Skip unknown fields
		}

		paramCount++
		query += fmt.Sprintf(", %s = $%d", dbField, paramCount)
		params = append(params, value)
	}

	query += fmt.Sprintf(" WHERE id = $1 RETURNING id")

	row := r.db.QueryRowContext(ctx, query, params...)
	var returnedID string
	if err := row.Scan(&returnedID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperrors.NotFound(fmt.Sprintf("user: %s", id))
		}
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// Search searches users by query
func (r *userRepository) Search(ctx context.Context, query string, limit int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	sqlQuery := `
		SELECT 
			id, email, username, first_name, last_name, bio, 
			auth_provider, image_url, is_active, created_at, updated_at
		FROM 
			"users"
		WHERE 
			username ILIKE $1 OR
			first_name ILIKE $1 OR
			last_name ILIKE $1
		LIMIT $2
	`

	searchPattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, sqlQuery, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var user model.User
		var firstName, lastName, bio, imageURL, email sql.NullString
		var updatedAt sql.NullTime
		var authProviderStr string

		err := rows.Scan(
			&user.ID,
			&email,
			&user.Username,
			&firstName,
			&lastName,
			&bio,
			&authProviderStr,
			&imageURL,
			&user.IsActive,
			&user.CreatedAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}

		// Handle null fields
		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if bio.Valid {
			user.Bio = &bio.String
		}
		if imageURL.Valid {
			user.ImageURL = &imageURL.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = &updatedAt.Time
		}

		user.AuthProvider = model.AuthProvider(authProviderStr)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return users, nil
}

// AddFollow creates a follow relationship
func (r *userRepository) AddFollow(ctx context.Context, followerID, followingID string) error {
	// Check if users exist first
	for _, id := range []string{followerID, followingID} {
		exists, err := r.userExists(ctx, id)
		if err != nil {
			return err
		}
		if !exists {
			return apperrors.NotFound(fmt.Sprintf("user: %s", id))
		}
	}

	// Check if already following
	isFollowing, err := r.IsFollowing(ctx, followerID, followingID)
	if err != nil {
		return err
	}
	if isFollowing {
		return nil // Already following, just return success
	}

	// Add follow
	query := `INSERT INTO follows (follower_id, following_id, created_at) VALUES ($1, $2, $3)`
	_, err = r.db.ExecContext(ctx, query, followerID, followingID, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("add follow: %w", err)
	}

	return nil
}

// RemoveFollow removes a follow relationship
func (r *userRepository) RemoveFollow(ctx context.Context, followerID, followingID string) error {
	query := `DELETE FROM follows WHERE follower_id = $1 AND following_id = $2`
	_, err := r.db.ExecContext(ctx, query, followerID, followingID)
	if err != nil {
		return fmt.Errorf("remove follow: %w", err)
	}

	return nil
}

// IsFollowing checks if a user is following another
func (r *userRepository) IsFollowing(ctx context.Context, followerID, followingID string) (bool, error) {
	query := `SELECT COUNT(*) FROM follows WHERE follower_id = $1 AND following_id = $2`

	var count int
	err := r.db.QueryRowContext(ctx, query, followerID, followingID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check following: %w", err)
	}

	return count > 0, nil
}

// GetFollowerCount returns the number of followers for a user
func (r *userRepository) GetFollowerCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM follows WHERE following_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get follower count: %w", err)
	}

	return count, nil
}

// GetFollowingCount returns the number of users a user is following
func (r *userRepository) GetFollowingCount(ctx context.Context, userID string) (int, error) {
	query := `SELECT COUNT(*) FROM follows WHERE follower_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get following count: %w", err)
	}

	return count, nil
}

// GetFollowers returns a list of users following the given user
func (r *userRepository) GetFollowers(ctx context.Context, userID string, limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			u.id, u.email, u.username, u.first_name, u.last_name, u.bio, 
			u.auth_provider, u.image_url, u.is_active, u.created_at, u.updated_at
		FROM 
			"users" u
		JOIN 
			follows f ON u.id = f.follower_id
		WHERE 
			f.following_id = $1
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get followers: %w", err)
	}
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var user model.User
		var firstName, lastName, bio, imageURL, email sql.NullString
		var updatedAt sql.NullTime
		var authProviderStr string

		err := rows.Scan(
			&user.ID,
			&email,
			&user.Username,
			&firstName,
			&lastName,
			&bio,
			&authProviderStr,
			&imageURL,
			&user.IsActive,
			&user.CreatedAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan follower row: %w", err)
		}

		// Handle null fields
		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if bio.Valid {
			user.Bio = &bio.String
		}
		if imageURL.Valid {
			user.ImageURL = &imageURL.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = &updatedAt.Time
		}

		user.AuthProvider = model.AuthProvider(authProviderStr)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return users, nil
}

// GetFollowing returns a list of users the given user is following
func (r *userRepository) GetFollowing(ctx context.Context, userID string, limit, offset int) ([]*model.User, error) {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	query := `
		SELECT 
			u.id, u.email, u.username, u.first_name, u.last_name, u.bio, 
			u.auth_provider, u.image_url, u.is_active, u.created_at, u.updated_at
		FROM 
			"users" u
		JOIN 
			follows f ON u.id = f.following_id
		WHERE 
			f.follower_id = $1
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get following: %w", err)
	}
	defer rows.Close()

	var users []*model.User

	for rows.Next() {
		var user model.User
		var firstName, lastName, bio, imageURL, email sql.NullString
		var updatedAt sql.NullTime
		var authProviderStr string

		err := rows.Scan(
			&user.ID,
			&email,
			&user.Username,
			&firstName,
			&lastName,
			&bio,
			&authProviderStr,
			&imageURL,
			&user.IsActive,
			&user.CreatedAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scan following row: %w", err)
		}

		// Handle null fields
		if email.Valid {
			user.Email = email.String
		}
		if firstName.Valid {
			user.FirstName = &firstName.String
		}
		if lastName.Valid {
			user.LastName = &lastName.String
		}
		if bio.Valid {
			user.Bio = &bio.String
		}
		if imageURL.Valid {
			user.ImageURL = &imageURL.String
		}
		if updatedAt.Valid {
			user.UpdatedAt = &updatedAt.Time
		}

		user.AuthProvider = model.AuthProvider(authProviderStr)

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration: %w", err)
	}

	return users, nil
}

// Helper function to check if a user exists
func (r *userRepository) userExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM "users" WHERE id = $1)`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check user exists: %w", err)
	}
	return exists, nil
}
