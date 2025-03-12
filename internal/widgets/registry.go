package widgets

import (
	"encoding/json"
	"fmt"
)

// WidgetTypeConfig defines configuration for a widget type
type WidgetTypeConfig struct {
	Type             string                 `json:"type"`
	DisplayName      string                 `json:"displayName"`
	Description      string                 `json:"description"`
	DefaultComponent string                 `json:"defaultComponent"`
	DefaultW         int                    `json:"defaultW"`
	DefaultH         int                    `json:"defaultH"`
	MinW             int                    `json:"minW"`
	MinH             int                    `json:"minH"`
	MaxW             int                    `json:"maxW"`
	MaxH             int                    `json:"maxH"`
	Schema           map[string]interface{} `json:"schema"`
	DefaultSettings  map[string]interface{} `json:"defaultSettings"`
	Customizations   []string               `json:"customizations"`
}

// Registry holds all available widget types
var Registry = map[string]WidgetTypeConfig{
	"experience": {
		Type:             "experience",
		DisplayName:      "Work Experience",
		Description:      "Showcase your professional experience",
		DefaultComponent: "ExperienceWidget",
		DefaultW:         6,
		DefaultH:         4,
		MinW:             3,
		MinH:             2,
		MaxW:             12,
		MaxH:             8,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"experiences": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"company":     map[string]interface{}{"type": "string"},
							"title":       map[string]interface{}{"type": "string"},
							"startDate":   map[string]interface{}{"type": "string"},
							"endDate":     map[string]interface{}{"type": "string"},
							"description": map[string]interface{}{"type": "string"},
						},
						"required": []string{"company", "title", "startDate"},
					},
				},
				"showDates": map[string]interface{}{
					"type":    "boolean",
					"default": true,
				},
			},
		},
		DefaultSettings: map[string]interface{}{
			"experiences": []interface{}{},
			"showDates":   true,
		},
		Customizations: []string{"backgroundColor", "borderRadius", "showTitle"},
	},
	"education": {
		Type:             "education",
		DisplayName:      "Education",
		Description:      "Showcase your educational background",
		DefaultComponent: "EducationWidget",
		DefaultW:         6,
		DefaultH:         3,
		MinW:             3,
		MinH:             2,
		MaxW:             12,
		MaxH:             6,
		Schema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"schools": map[string]interface{}{
					"type": "array",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"institution": map[string]interface{}{"type": "string"},
							"degree":      map[string]interface{}{"type": "string"},
							"field":       map[string]interface{}{"type": "string"},
							"startDate":   map[string]interface{}{"type": "string"},
							"endDate":     map[string]interface{}{"type": "string"},
						},
						"required": []string{"institution", "startDate"},
					},
				},
			},
		},
		DefaultSettings: map[string]interface{}{
			"schools": []interface{}{},
		},
		Customizations: []string{"backgroundColor", "borderRadius", "showTitle"},
	},
	// Add more widget types as needed
}

// GetWidgetTypeConfig returns the configuration for a widget type
func GetWidgetTypeConfig(widgetType string) (WidgetTypeConfig, error) {
	config, exists := Registry[widgetType]
	if !exists {
		return WidgetTypeConfig{}, fmt.Errorf("unknown widget type: %s", widgetType)
	}
	return config, nil
}

// GetWidgetTypes returns all available widget types
func GetWidgetTypes() map[string]WidgetTypeConfig {
	return Registry
}

// ValidWidgetType checks if a widget type is valid
func ValidWidgetType(widgetType string) bool {
	_, exists := Registry[widgetType]
	return exists
}

// GetDefaultSettings returns default settings for a widget type
func GetDefaultSettings(widgetType string) (string, error) {
	config, err := GetWidgetTypeConfig(widgetType)
	if err != nil {
		return "", err
	}

	settings, err := json.Marshal(config.DefaultSettings)
	if err != nil {
		return "", err
	}

	return string(settings), nil
}
