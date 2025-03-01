package model

import (
	"time"
)

// Visibility type for posts
type Visibility string

const (
	VisibilityPublic  Visibility = "public"
	VisibilityPrivate Visibility = "private"
)

// Post represents a user post
type Post struct {
	ID         string     `json:"id"`
	UserID     string     `json:"userId"`
	Content    string     `json:"content"`
	EmbedURLs  []string   `json:"embedUrls,omitempty"`
	Hashtags   []string   `json:"hashtags,omitempty"`
	Visibility Visibility `json:"visibility"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`

	// Optional joined fields
	Author    *User      `json:"author,omitempty"`
	Reactions []Reaction `json:"reactions,omitempty"`
}

// CreatePostRequest is used when creating a new post
type CreatePostRequest struct {
	Content   string   `json:"content" validate:"required,max=500"`
	EmbedURLs []string `json:"embedUrls,omitempty" validate:"omitempty,dive,url,max=3"`
	Hashtags  []string `json:"hashtags,omitempty" validate:"omitempty,dive,max=30,max=5"`
}

// UpdatePostRequest is used when updating an existing post
type UpdatePostRequest struct {
	Content   string   `json:"content" validate:"required,max=500"`
	EmbedURLs []string `json:"embedUrls,omitempty" validate:"omitempty,dive,url,max=3"`
	Hashtags  []string `json:"hashtags,omitempty" validate:"omitempty,dive,max=30,max=5"`
}

// FeedRequest is used to get a user's feed
type FeedRequest struct {
	Limit  int `json:"limit,omitempty"`
	Offset int `json:"offset,omitempty"`
}

// FeedResponse is the response for a feed request
type FeedResponse struct {
	Posts []Post `json:"posts"`
	Total int    `json:"total"`
}
