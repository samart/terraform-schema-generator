package validator

import (
	"encoding/json"
	"fmt"

	"github.com/samart/terraform-schema-generator/pkg/converter"
)

// Validator validates JSON Schema 7 documents
type Validator struct{}

// NewValidator creates a new validator instance
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateSchema validates that the generated schema is valid JSON Schema 7
func (v *Validator) ValidateSchema(schema *converter.JSONSchema7) error {
	// Check required fields
	if schema.Schema == "" {
		return fmt.Errorf("missing $schema field")
	}

	if schema.Schema != "http://json-schema.org/draft-07/schema#" {
		return fmt.Errorf("invalid schema version: expected draft-07, got %s", schema.Schema)
	}

	if schema.Type == "" {
		return fmt.Errorf("missing type field")
	}

	// Validate type is valid
	validTypes := map[string]bool{
		"object":  true,
		"array":   true,
		"string":  true,
		"number":  true,
		"integer": true,
		"boolean": true,
		"null":    true,
	}

	if !validTypes[schema.Type] {
		return fmt.Errorf("invalid type: %s", schema.Type)
	}

	// Validate properties
	if schema.Type == "object" && len(schema.Properties) == 0 {
		return fmt.Errorf("object type must have at least one property")
	}

	// Validate required fields exist in properties
	for _, required := range schema.Required {
		if _, exists := schema.Properties[required]; !exists {
			return fmt.Errorf("required field '%s' not found in properties", required)
		}
	}

	// Validate that the schema can be marshaled to JSON
	_, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("schema cannot be marshaled to JSON: %w", err)
	}

	return nil
}

// ValidateProperty validates a single property definition
func (v *Validator) ValidateProperty(name string, prop converter.Property) error {
	// Type validation
	if prop.Type == nil {
		return fmt.Errorf("property '%s' missing type", name)
	}

	// Validate numeric constraints
	if prop.Minimum != nil && prop.Maximum != nil {
		if *prop.Minimum > *prop.Maximum {
			return fmt.Errorf("property '%s': minimum (%f) cannot be greater than maximum (%f)",
				name, *prop.Minimum, *prop.Maximum)
		}
	}

	// Validate string constraints
	if prop.MinLength != nil && prop.MaxLength != nil {
		if *prop.MinLength > *prop.MaxLength {
			return fmt.Errorf("property '%s': minLength (%d) cannot be greater than maxLength (%d)",
				name, *prop.MinLength, *prop.MaxLength)
		}
	}

	// Validate pattern is valid regex if present
	if prop.Pattern != "" { //nolint:staticcheck // Placeholder for future regex validation
		// In production, you'd compile and validate the regex pattern
		// regexp.MustCompile(prop.Pattern)
	}

	return nil
}
