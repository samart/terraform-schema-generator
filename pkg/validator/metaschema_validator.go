package validator

import (
	"encoding/json"
	"fmt"

	"github.com/xeipuuv/gojsonschema"
)

// MetaSchemaValidator validates JSON schemas against the JSON Schema Draft 7 meta-schema
type MetaSchemaValidator struct {
	metaSchemaURL string
}

// NewMetaSchemaValidator creates a new meta-schema validator
func NewMetaSchemaValidator() *MetaSchemaValidator {
	return &MetaSchemaValidator{
		metaSchemaURL: "http://json-schema.org/draft-07/schema#",
	}
}

// ValidateAgainstMetaSchema validates a JSON schema against the JSON Schema Draft 7 meta-schema
func (v *MetaSchemaValidator) ValidateAgainstMetaSchema(schemaBytes []byte) error {
	// Load the JSON Schema Draft 7 meta-schema
	metaSchemaLoader := gojsonschema.NewReferenceLoader(v.metaSchemaURL)

	// Parse the schema to validate
	var schemaData interface{}
	if err := json.Unmarshal(schemaBytes, &schemaData); err != nil {
		return fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	// Create a document loader from the parsed schema
	documentLoader := gojsonschema.NewGoLoader(schemaData)

	// Validate the schema against the meta-schema
	result, err := gojsonschema.Validate(metaSchemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// Check if the schema is valid
	if !result.Valid() {
		errMsg := "schema is not valid against JSON Schema Draft 7 meta-schema:\n"
		for _, err := range result.Errors() {
			errMsg += fmt.Sprintf("  - %s\n", err.String())
		}
		return fmt.Errorf("%s", errMsg)
	}

	return nil
}

// ValidateAgainstMetaSchemaWithDetails validates and returns detailed validation information
func (v *MetaSchemaValidator) ValidateAgainstMetaSchemaWithDetails(schemaBytes []byte) (*ValidationResult, error) {
	metaSchemaLoader := gojsonschema.NewReferenceLoader(v.metaSchemaURL)

	var schemaData interface{}
	if err := json.Unmarshal(schemaBytes, &schemaData); err != nil {
		return nil, fmt.Errorf("failed to parse schema JSON: %w", err)
	}

	documentLoader := gojsonschema.NewGoLoader(schemaData)

	result, err := gojsonschema.Validate(metaSchemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	validationResult := &ValidationResult{
		Valid:  result.Valid(),
		Errors: []string{},
	}

	if !result.Valid() {
		for _, err := range result.Errors() {
			validationResult.Errors = append(validationResult.Errors, err.String())
		}
	}

	return validationResult, nil
}

// ValidationResult represents the result of a meta-schema validation
type ValidationResult struct {
	Valid  bool
	Errors []string
}
