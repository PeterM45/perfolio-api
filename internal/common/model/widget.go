package model

import (
	"encoding/json"
	"time"
)

// Widget represents a user profile widget
type Widget struct {
	ID          string     `json:"id"`
	UserID      string     `json:"userId"`
	Type        string     `json:"type"`
	Component   string     `json:"component"`
	X           int        `json:"x"`
	Y           int        `json:"y"`
	W           int        `json:"w"`
	H           int        `json:"h"`
	Settings    *string    `json:"settings,omitempty"`
	DisplayName *string    `json:"displayName,omitempty"`
	IsVisible   bool       `json:"isVisible"`
	Version     int        `json:"version"`
	CreatedAt   time.Time  `json:"createdAt,omitempty"`
	UpdatedAt   time.Time  `json:"updatedAt,omitempty"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty"`
}

// WidgetType represents a widget type configuration
type WidgetType struct {
	Type             string                 `json:"type"`
	DisplayName      string                 `json:"displayName"`
	Description      string                 `json:"description"`
	DefaultComponent string                 `json:"defaultComponent"`
	DefaultSize      WidgetSize             `json:"defaultSize"`
	MinSize          WidgetSize             `json:"minSize"`
	MaxSize          WidgetSize             `json:"maxSize"`
	Schema           map[string]interface{} `json:"schema"`
	DefaultSettings  map[string]interface{} `json:"defaultSettings"`
	Customizations   []string               `json:"customizations"`
}

// WidgetSize represents a widget's size
type WidgetSize struct {
	W int `json:"w"`
	H int `json:"h"`
}

// CreateWidgetRequest is used to create a new widget
type CreateWidgetRequest struct {
	Type        string `json:"type" validate:"required"`
	Component   string `json:"component" validate:"required"`
	X           int    `json:"x" validate:"min=0"`
	Y           int    `json:"y" validate:"min=0"`
	W           int    `json:"w" validate:"required,min=1,max=12"`
	H           int    `json:"h" validate:"required,min=1"`
	Settings    string `json:"settings,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// UpdateWidgetRequest is used to update an existing widget
type UpdateWidgetRequest struct {
	Type        *string `json:"type,omitempty"`
	Component   *string `json:"component,omitempty"`
	X           *int    `json:"x,omitempty" validate:"omitempty,min=0"`
	Y           *int    `json:"y,omitempty" validate:"omitempty,min=0"`
	W           *int    `json:"w,omitempty" validate:"omitempty,min=1,max=12"`
	H           *int    `json:"h,omitempty" validate:"omitempty,min=1"`
	Settings    *string `json:"settings,omitempty"`
	DisplayName *string `json:"displayName,omitempty"`
	IsVisible   *bool   `json:"isVisible,omitempty"`
	Version     *int    `json:"version,omitempty" validate:"required"`
}

// WidgetPositionUpdate is used to update widget positions in batch
type WidgetPositionUpdate struct {
	ID      string `json:"id" validate:"required"`
	UserID  string `json:"userId" validate:"required"`
	X       int    `json:"x" validate:"min=0"`
	Y       int    `json:"y" validate:"min=0"`
	W       int    `json:"w" validate:"min=1,max=12"`
	H       int    `json:"h" validate:"min=1"`
	Version int    `json:"version" validate:"required"`
}

// BatchUpdateWidgetsRequest is used for batch position updates
type BatchUpdateWidgetsRequest struct {
	Updates []WidgetPositionUpdate `json:"updates" validate:"required,dive"`
}

// GetSettingsMap parses the widget settings into a map
func (w *Widget) GetSettingsMap() (map[string]interface{}, error) {
	if w.Settings == nil {
		return map[string]interface{}{}, nil
	}

	var settings map[string]interface{}
	err := json.Unmarshal([]byte(*w.Settings), &settings)
	return settings, err
}
