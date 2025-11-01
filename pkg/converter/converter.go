package converter

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/samart/terraform-schema-generator/pkg/parser"
)

// JSONSchema7 represents a JSON Schema Draft 7 document
type JSONSchema7 struct {
	Schema      string                 `json:"$schema"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Type        string                 `json:"type"`
	Properties  map[string]Property    `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Definitions map[string]interface{} `json:"definitions,omitempty"`
}

// Property represents a JSON Schema property
type Property struct {
	Type        interface{} `json:"type,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Format      string      `json:"format,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	MinLength   *int        `json:"minLength,omitempty"`
	MaxLength   *int        `json:"maxLength,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
	Items       *Property   `json:"items,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Properties  interface{} `json:"properties,omitempty"`
	WriteOnly   bool        `json:"writeOnly,omitempty"`
}

// Converter converts Terraform variables to JSON Schema 7
type Converter struct{}

// NewConverter creates a new converter instance
func NewConverter() *Converter {
	return &Converter{}
}

// ConvertToJSONSchema7 converts parsed Terraform variables to JSON Schema 7
func (c *Converter) ConvertToJSONSchema7(parseResult *parser.ParseResult) (*JSONSchema7, error) {
	if len(parseResult.Variables) == 0 {
		return nil, fmt.Errorf("no variables to convert")
	}

	schema := &JSONSchema7{
		Schema:      "http://json-schema.org/draft-07/schema#",
		Title:       "Terraform Variables Schema",
		Description: "Generated JSON Schema from Terraform variable definitions",
		Type:        "object",
		Properties:  make(map[string]Property),
		Required:    []string{},
		Definitions: make(map[string]interface{}),
	}

	for _, variable := range parseResult.Variables {
		property := c.convertVariable(variable)
		schema.Properties[variable.Name] = property

		if variable.Required {
			schema.Required = append(schema.Required, variable.Name)
		}
	}

	return schema, nil
}

// convertVariable converts a single Terraform variable to a JSON Schema property
func (c *Converter) convertVariable(variable parser.Variable) Property {
	property := Property{
		Description: variable.Description,
		Default:     variable.Default,
		WriteOnly:   variable.Sensitive,
	}

	// Map Terraform types to JSON Schema types
	property.Type = c.mapTerraformTypeToJSONSchema(variable.Type)

	// Handle validation rules
	c.applyValidationRules(&property, variable.Validation)

	return property
}

// mapTerraformTypeToJSONSchema maps Terraform types to JSON Schema types
func (c *Converter) mapTerraformTypeToJSONSchema(tfType string) interface{} {
	// Handle primitive types
	switch strings.ToLower(tfType) {
	case "string":
		return "string"
	case "number":
		return "number"
	case "bool", "boolean":
		return "boolean"
	case "any":
		return []string{"string", "number", "boolean", "object", "array", "null"}
	}

	// Handle collection types
	if strings.HasPrefix(tfType, "list") {
		return "array"
	}

	if strings.HasPrefix(tfType, "set") {
		return "array"
	}

	if strings.HasPrefix(tfType, "map") {
		return "object"
	}

	if strings.HasPrefix(tfType, "object") {
		return "object"
	}

	if strings.HasPrefix(tfType, "tuple") {
		return "array"
	}

	// Default to string if type cannot be determined
	return "string"
}

// applyValidationRules applies Terraform validation rules to JSON Schema properties
func (c *Converter) applyValidationRules(property *Property, rules []parser.ValidationRule) {
	for _, rule := range rules {
		// Parse common validation patterns
		condition := strings.ToLower(rule.Condition)

		// Length validations for strings
		if strings.Contains(condition, "length") {
			if strings.Contains(condition, ">=") || strings.Contains(condition, "min") {
				minLen := 1
				property.MinLength = &minLen
			}
			if strings.Contains(condition, "<=") || strings.Contains(condition, "max") {
				maxLen := 255
				property.MaxLength = &maxLen
			}
		}

		// Pattern validations
		if strings.Contains(condition, "regex") || strings.Contains(condition, "can(regex") {
			// Extract regex pattern if possible
			// This is simplified - production code would parse the actual regex
			property.Pattern = ".*"
		}

		// Enum validations
		if strings.Contains(condition, "contains") {
			// This would require parsing the actual values
			// Simplified for now
		}
	}
}

// ToJSON converts the schema to JSON bytes
func (c *Converter) ToJSON(schema *JSONSchema7) ([]byte, error) {
	return json.MarshalIndent(schema, "", "  ")
}

// ToJSONString converts the schema to a JSON string
func (c *Converter) ToJSONString(schema *JSONSchema7) (string, error) {
	bytes, err := c.ToJSON(schema)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
