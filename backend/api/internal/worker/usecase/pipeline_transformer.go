package usecase

import (
	"context"
	"strings"

	"omnilogs-api/models"
)

// PipelineTransformer applies transformations defined in ConfigJSON to a single field value.
func PipelineTransformer(ctx context.Context, field models.LogFieldDefinition, value any) (any, error) {
	if value == nil || field.ConfigJSON == nil {
		return value, nil
	}

	// Handle string transformations
	if str, ok := value.(string); ok {
		// Specific type configs might have shorthand like "Trim", "Uppercase"
		if field.ConfigJSON.StringConfig != nil {
			if field.ConfigJSON.StringConfig.Trim {
				str = strings.TrimSpace(str)
			}
			if field.ConfigJSON.StringConfig.Uppercase {
				str = strings.ToUpper(str)
			}
			if field.ConfigJSON.StringConfig.Lowercase {
				str = strings.ToLower(str)
			}
		}
		
		// Full transformation pipeline support
		for _, step := range field.ConfigJSON.Transform {
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
