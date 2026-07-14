package usecase

import (
	"context"
	"encoding/json"
	"strings"

	"omnilogs-api/models"
)

// PipelineTransformer applies transformations defined in ConfigJSON to a single field value.
func PipelineTransformer(ctx context.Context, field models.LogFieldDefinition, value any) (any, error) {
	if value == nil || field.ConfigJSON == nil {
		return value, nil
	}

	var config models.FieldConfigJSON
	if err := json.Unmarshal(*field.ConfigJSON, &config); err != nil {
		return value, nil
	}

	// Handle string transformations
	if str, ok := value.(string); ok {
		// Specific type configs might have shorthand like "Trim", "Uppercase"
		if config.StringConfig != nil {
			if config.StringConfig.Trim {
				str = strings.TrimSpace(str)
			}
			if config.StringConfig.Uppercase {
				str = strings.ToUpper(str)
			}
			if config.StringConfig.Lowercase {
				str = strings.ToLower(str)
			}
		}
		
		// Full transformation pipeline support
		for _, step := range config.Transform {
			switch step.Type {
			case "TRIM":
				str = strings.TrimSpace(str)
			case "UPPERCASE":
				str = strings.ToUpper(str)
			case "LOWERCASE":
				str = strings.ToLower(str)
			}
		}
		
		return str, nil
	}

	return value, nil
}
