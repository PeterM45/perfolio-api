package model

import (
	"time"
)

// AuthProvider represents authentication provider types
type AuthProvider string

const (
	AuthProviderCustom AuthProvider = "custom" // New custom auth provider
	AuthProviderOAuth  AuthProvider = "oauth"
	AuthProviderEmail  AuthProvider = "email"
	AuthProviderGoogle AuthProvider = "google"
	AuthProviderGithub AuthProvider = "github"
)

// User represents a user in the system
type User struct {
	ID           string       `json:"id"`
	Email        string       `json:"email,omitempty"`
	Username     string       `json:"username"`
	FirstName    *string      `json:"firstName,omitempty"`
	LastName     *string      `json:"lastName,omitempty"`
	Bio          *string      `json:"bio,omitempty"`
	AuthProvider AuthProvider `json:"authProvider"`
	PasswordHash string       `json:"-"` // Stored hashed password, not exposed in JSON
	ImageURL     *string      `json:"imageUrl,omitempty"`
	IsActive     bool         `json:"isActive"`
	CreatedAt    time.Time    `json:"createdAt"`
	UpdatedAt    *time.Time   `json:"updatedAt,omitempty"`
}

// CreateUserRequest is used when creating a new user
type CreateUserRequest struct {
	ID           string       `json:"id,omitempty"`                    // Made optional for custom auth
	Email        string       `json:"email" validate:"required,email"` // Required for custom auth
	Username     string       `json:"username" validate:"required,min=3,max=64"`
	Password     string       `json:"password,omitempty" validate:"omitempty,min=6"` // For custom auth
	PasswordHash string       `json:"-"`                                             // Internal use only, set by auth handler
	FirstName    *string      `json:"firstName,omitempty" validate:"omitempty,max=64"`
	LastName     *string      `json:"lastName,omitempty" validate:"omitempty,max=64"`
	Bio          *string      `json:"bio,omitempty" validate:"omitempty,max=500"`
	AuthProvider AuthProvider `json:"authProvider" validate:"required"`
	ImageURL     *string      `json:"imageUrl,omitempty" validate:"omitempty,url"`
}

// UpdateUserRequest is used when updating an existing user
type UpdateUserRequest struct {
	Username        *string `json:"username,omitempty" validate:"omitempty,min=3,max=64"`
	FirstName       *string `json:"firstName,omitempty" validate:"omitempty,max=64"`
	LastName        *string `json:"lastName,omitempty" validate:"omitempty,max=64"`
	Bio             *string `json:"bio,omitempty" validate:"omitempty,max=500"`
	ImageURL        *string `json:"imageUrl,omitempty" validate:"omitempty,url"`
	IsActive        *bool   `json:"isActive,omitempty"`
	Password        *string `json:"password,omitempty" validate:"omitempty,min=6"` // For updating password
	CurrentPassword *string `json:"currentPassword,omitempty"`                     // For password verification
}

// ChangePasswordRequest is used for password changes
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=6"`
}

// SearchUserRequest is used when searching for users
type SearchUserRequest struct {
	Query string `json:"query" validate:"required"`
	Limit int    `json:"limit,omitempty"`
}

// SearchUserResponse is the response for user search
type SearchUserResponse struct {
	Users []User `json:"users"`
}

// UserProfile is a public-facing user profile
type UserProfile struct {
	ID             string  `json:"id"`
	Username       string  `json:"username"`
	FirstName      *string `json:"firstName,omitempty"`
	LastName       *string `json:"lastName,omitempty"`
	Bio            *string `json:"bio,omitempty"`
	ImageURL       *string `json:"imageUrl,omitempty"`
	FollowerCount  int     `json:"followerCount"`
	FollowingCount int     `json:"followingCount"`
}

// UserResponse is the standard user response with follow stats
type UserResponse struct {
	User           User `json:"user"`
	FollowerCount  int  `json:"followerCount"`
	FollowingCount int  `json:"followingCount"`
	IsFollowing    bool `json:"isFollowing"`
}
