package model

import (
	"time"
)

// Widget represents a user profile widget
type Widget struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Type      string    `json:"type"`
	Component string    `json:"component"`
	X         int       `json:"x"`
	Y         int       `json:"y"`
	W         int       `json:"w"`
	H         int       `json:"h"`
	Settings  *string   `json:"settings,omitempty"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	Version   int       `json:"version,omitempty"`
}

// CreateWidgetRequest is used to create a new widget
type CreateWidgetRequest struct {
	Type      string `json:"type" validate:"required"`
	Component string `json:"component" validate:"required"`
	X         int    `json:"x" validate:"min=0"`
	Y         int    `json:"y" validate:"min=0"`
	W         int    `json:"w" validate:"required,min=1,max=12"`
	H         int    `json:"h" validate:"required,min=1"`
	Settings  string `json:"settings,omitempty"`
}

// UpdateWidgetRequest is used to update an existing widget
type UpdateWidgetRequest struct {
	Type      *string `json:"type,omitempty"`
	Component *string `json:"component,omitempty"`
	X         *int    `json:"x,omitempty" validate:"omitempty,min=0"`
	Y         *int    `json:"y,omitempty" validate:"omitempty,min=0"`
	W         *int    `json:"w,omitempty" validate:"omitempty,min=1,max=12"`
	H         *int    `json:"h,omitempty" validate:"omitempty,min=1"`
	Settings  *string `json:"settings,omitempty"`
}

// WidgetPositionUpdate is used to update widget positions in batch
type WidgetPositionUpdate struct {
	ID     string `json:"id" validate:"required"`
	UserID string `json:"userId" validate:"required"`
	X      int    `json:"x" validate:"min=0"`
	Y      int    `json:"y" validate:"min=0"`
	W      int    `json:"w" validate:"min=1,max=12"`
	H      int    `json:"h" validate:"min=1"`
}

// BatchUpdateWidgetsRequest is used for batch position updates
type BatchUpdateWidgetsRequest struct {
	Updates []WidgetPositionUpdate `json:"updates" validate:"required,dive"`
}
