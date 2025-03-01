package model

import (
	"time"
)

// ReactionType represents reaction types
type ReactionType string

const (
	ReactionTypeLike       ReactionType = "like"
	ReactionTypeCelebrate  ReactionType = "celebrate"
	ReactionTypeSupport    ReactionType = "support"
	ReactionTypeInsightful ReactionType = "insightful"
	ReactionTypeCurious    ReactionType = "curious"
)

// Reaction represents a reaction to a post
type Reaction struct {
	ID        string       `json:"id"`
	PostID    string       `json:"postId"`
	UserID    string       `json:"userId"`
	Type      ReactionType `json:"type"`
	CreatedAt time.Time    `json:"createdAt"`

	// Optional joined fields
	User *User `json:"user,omitempty"`
	Post *Post `json:"post,omitempty"`
}

// ReactionRequest is used to react to a post
type ReactionRequest struct {
	PostID string       `json:"postId" validate:"required"`
	Type   ReactionType `json:"type" validate:"required,oneof=like celebrate support insightful curious"`
}
