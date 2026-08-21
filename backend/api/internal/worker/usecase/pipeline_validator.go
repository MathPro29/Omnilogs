package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"omnilogs-api/models"
)

// PipelineValidator validates a single custom field value against its schema configuration.
func PipelineValidator(ctx context.Context, field models.LogFieldDefinition, value any) error {
	var config models.FieldConfigJSON
	if field.ConfigJSON != nil {
		if err := json.Unmarshal(*field.ConfigJSON, &config); err != nil {
			return fmt.Errorf("field %s has invalid config_json: %w", field.FieldKey, err)
		}
	}

	valConfig := config.Validation
	if (field.IsRequired || (valConfig != nil && valConfig.Required)) && value == nil {
		return fmt.Errorf("field %s is required", field.FieldKey)
	}
	if value == nil {
		if valConfig != nil && !valConfig.Nullable {
			return fmt.Errorf("field %s cannot be null", field.FieldKey)
		}
		return nil
	}

	if config.StringConfig != nil {
		if str, ok := value.(string); ok {
			strConfig := config.StringConfig
			if strConfig.MinLength != nil && len(str) < *strConfig.MinLength {
				return fmt.Errorf("field %s must be at least %d characters", field.FieldKey, *strConfig.MinLength)
			}
			if strConfig.MaxLength != nil && len(str) > *strConfig.MaxLength {
				return fmt.Errorf("field %s must be at most %d characters", field.FieldKey, *strConfig.MaxLength)
			}
			if strConfig.Regex != nil && *strConfig.Regex != "" {
				expression, err := regexp.Compile(*strConfig.Regex)
				if err != nil {
					return fmt.Errorf("field %s has invalid regex: %w", field.FieldKey, err)
				}
				if !expression.MatchString(str) {
					return fmt.Errorf("field %s does not match configured pattern", field.FieldKey)
				}
			}
		}
	}

	if config.NumberConfig != nil {
		var num float64
		isNum := false
		switch v := value.(type) {
		case float64:
			num, isNum = v, true
		case int:
			num, isNum = float64(v), true
		case int32:
			num, isNum = float64(v), true
		case int64:
			num, isNum = float64(v), true
		}
		if isNum {
			numConfig := config.NumberConfig
			if numConfig.Min != nil && num < *numConfig.Min {
				return fmt.Errorf("field %s must be at least %f", field.FieldKey, *numConfig.Min)
			}
			if numConfig.Max != nil && num > *numConfig.Max {
				return fmt.Errorf("field %s must be at most %f", field.FieldKey, *numConfig.Max)
			}
		}
	}

	return nil
}
