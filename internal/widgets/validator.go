package widgets

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// ValidateWidgetSettings validates widget settings against the schema
func ValidateWidgetSettings(widgetType string, settings string) error {
	config, err := GetWidgetTypeConfig(widgetType)
	if err != nil {
		return err
	}

	// If no settings provided, return early
	if settings == "" {
		return nil
	}

	var settingsObj map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &settingsObj); err != nil {
		return fmt.Errorf("invalid settings JSON: %w", err)
	}

	schemaLoader := gojsonschema.NewGoLoader(config.Schema)
	documentLoader := gojsonschema.NewGoLoader(settingsObj)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		var errors string
		for _, err := range result.Errors() {
			errors += fmt.Sprintf("- %s\n", err.String())
		}
		return fmt.Errorf("invalid settings: \n%s", errors)
	}

	return nil
}

// ValidateCustomization checks if a customization is allowed for a widget type
func ValidateCustomization(widgetType string, customization string) bool {
	config, err := GetWidgetTypeConfig(widgetType)
	if err != nil {
		return false
	}

	for _, allowed := range config.Customizations {
		if allowed == customization {
			return true
		}
	}

	return false
}
