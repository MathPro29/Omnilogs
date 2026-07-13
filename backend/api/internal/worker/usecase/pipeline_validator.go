package usecase

import (
	"context"
	"fmt"
	"regexp"

	"omnilogs-api/models"
)

// PipelineValidator validates a single custom field value against its schema configuration.
func PipelineValidator(ctx context.Context, field models.LogFieldDefinition, value any) error {
	if field.ConfigJSON == nil || field.ConfigJSON.Validation == nil {
		return nil
	}

	valConfig := field.ConfigJSON.Validation

	if valConfig.Required && value == nil {
		return fmt.Errorf("field %s is required", field.FieldKey)
	}

	if value == nil && !valConfig.Nullable {
		// Depending on strictness, we might error if it's nil and not nullable
		return fmt.Errorf("field %s cannot be null", field.FieldKey)
	}

	if value == nil {
		return nil // Nothing else to validate if nil
	}

	switch field.FieldType {
	case nil:
		// Fallback for old schema
		return nil
	}

	if field.ConfigJSON.StringConfig != nil {
		if str, ok := value.(string); ok {
			strConfig := field.ConfigJSON.StringConfig
			if strConfig.MinLength != nil && len(str) < *strConfig.MinLength {
				return fmt.Errorf("field %s must be at least %d characters", field.FieldKey, *strConfig.MinLength)
			}
			if strConfig.MaxLength != nil && len(str) > *strConfig.MaxLength {
				return fmt.Errorf("field %s must be at most %d characters", field.FieldKey, *strConfig.MaxLength)
			}
			if strConfig.Regex != nil && *strConfig.Regex != "" {
				matched, err := regexp.MatchString(*strConfig.Regex, str)
				if err == nil && !matched {
					return fmt.Errorf("field %s does not match pattern %s", field.FieldKey, *strConfig.Regex)
				}
			}
		}
	}

	if field.ConfigJSON.NumberConfig != nil {
		var num float64
		isNum := false
		switch v := value.(type) {
		case float64:
			num = v
			isNum = true
		case int:
			num = float64(v)
			isNum = true
		case int32:
			num = float64(v)
			isNum = true
		case int64:
			num = float64(v)
			isNum = true
		}
		if isNum {
			numConfig := field.ConfigJSON.NumberConfig
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
