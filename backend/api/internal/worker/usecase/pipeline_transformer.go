package usecase

import (
	"context"
	"encoding/json"
	"fmt"
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
		return value, fmt.Errorf("field %s has invalid config_json: %w", field.FieldKey, err)
	}

	str, ok := value.(string)
	if !ok {
		return value, nil
	}

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

	for _, step := range config.Transform {
		switch strings.ToUpper(strings.TrimSpace(step.Type)) {
		case "TRIM":
			str = strings.TrimSpace(str)
		case "UPPERCASE":
			str = strings.ToUpper(str)
		case "LOWERCASE":
			str = strings.ToLower(str)
		case "":
			return value, fmt.Errorf("field %s contains an empty transformation type", field.FieldKey)
		default:
			return value, fmt.Errorf("field %s contains unsupported transformation %q", field.FieldKey, step.Type)
		}
	}

	return str, nil
}
