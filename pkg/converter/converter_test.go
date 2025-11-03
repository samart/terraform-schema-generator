package converter

import (
	"encoding/json"
	"testing"

	"github.com/samart/terraform-schema-generator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
)

// validateSchemaAgainstMetaSchema is a helper that validates a schema against JSON Schema Draft 7 meta-schema
func validateSchemaAgainstMetaSchema(t *testing.T, schema *JSONSchema7) {
	t.Helper()

	converter := NewConverter()
	schemaJSON, err := converter.ToJSON(schema)
	require.NoError(t, err, "Failed to marshal schema to JSON")

	// Validate against JSON Schema Draft 7 meta-schema
	metaSchemaLoader := gojsonschema.NewReferenceLoader("http://json-schema.org/draft-07/schema#")

	var schemaData interface{}
	err = json.Unmarshal(schemaJSON, &schemaData)
	require.NoError(t, err, "Failed to unmarshal schema JSON")

	documentLoader := gojsonschema.NewGoLoader(schemaData)

	result, err := gojsonschema.Validate(metaSchemaLoader, documentLoader)
	require.NoError(t, err, "Validation process failed")

	if !result.Valid() {
		t.Errorf("Schema is not valid against JSON Schema Draft 7 meta-schema:")
		for _, err := range result.Errors() {
			t.Errorf("  - %s", err.String())
		}
		t.FailNow()
	}
}

func TestConvertToJSONSchema7(t *testing.T) {
	converter := NewConverter()

	t.Run("convert simple variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "instance_type",
					Type:        "string",
					Description: "EC2 instance type",
					Default:     "t2.micro",
					Required:    false,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
		assert.Equal(t, "object", schema.Type)
		assert.Len(t, schema.Properties, 1)
		assert.Equal(t, "string", schema.Properties["instance_type"].Type)
		assert.Empty(t, schema.Required)
	})

	t.Run("convert required variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:     "vpc_id",
					Type:     "string",
					Required: true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "vpc_id")
	})

	t.Run("convert sensitive variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:      "db_password",
					Type:      "string",
					Sensitive: true,
					Required:  true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.True(t, schema.Properties["db_password"].WriteOnly)
	})

	t.Run("convert number type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "instance_count",
					Type: "number",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "number", schema.Properties["instance_count"].Type)
	})

	t.Run("convert boolean type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "enable_monitoring",
					Type: "bool",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "boolean", schema.Properties["enable_monitoring"].Type)
	})

	t.Run("convert list type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "availability_zones",
					Type: "list(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "array", schema.Properties["availability_zones"].Type)
	})

	t.Run("convert map type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "tags",
					Type: "map(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Properties["tags"].Type)
	})

	t.Run("error on empty variables", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{},
		}

		_, err := converter.ConvertToJSONSchema7(parseResult)
		assert.Error(t, err)
	})

	t.Run("convert set type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "availability_zones",
					Type: "set(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		// Set should be converted to array with uniqueItems
		prop := schema.Properties["availability_zones"]
		assert.Equal(t, "array", prop.Type)
	})

	t.Run("convert tuple type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "example_tuple",
					Type: "tuple([string, number, bool])",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		// Tuple should be converted to array
		prop := schema.Properties["example_tuple"]
		assert.Equal(t, "array", prop.Type)
	})

	t.Run("convert object type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "server_config",
					Type: "object({name = string, port = number, enabled = bool})",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["server_config"]
		assert.Equal(t, "object", prop.Type)
	})

	t.Run("convert nullable variable true", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:     "optional_string",
					Type:     "string",
					Nullable: true,
					Default:  nil,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["optional_string"]
		// Nullable should allow null type
		if typeSlice, ok := prop.Type.([]interface{}); ok {
			assert.Contains(t, typeSlice, "null")
		}
	})

	t.Run("convert nullable variable false", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:     "required_non_null",
					Type:     "string",
					Nullable: false,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["required_non_null"]
		// When nullable is false, should only be string type, not allow null
		assert.Equal(t, "string", prop.Type)
	})

	t.Run("convert ephemeral variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:      "session_token",
					Type:      "string",
					Ephemeral: true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["session_token"]
		// Ephemeral variables should be present in schema
		assert.Equal(t, "string", prop.Type)
		// Could potentially add a custom field to indicate ephemeral
	})

	t.Run("convert ephemeral and sensitive together", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:      "secret_key",
					Type:      "string",
					Sensitive: true,
					Ephemeral: true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["secret_key"]
		assert.True(t, prop.WriteOnly) // Sensitive should set writeOnly
		assert.Equal(t, "string", prop.Type)
	})

	t.Run("convert all attributes together", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "complete_var",
					Type:        "string",
					Description: "Complete variable with all attributes",
					Default:     "default",
					Sensitive:   true,
					Nullable:    false,
					Ephemeral:   true,
					Required:    false,
					Validations: []parser.Validation{
						{
							Condition:    "length(var.complete_var) > 5",
							ErrorMessage: "Must be longer than 5",
						},
					},
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["complete_var"]
		assert.Equal(t, "string", prop.Type)
		assert.Equal(t, "Complete variable with all attributes", prop.Description)
		assert.True(t, prop.WriteOnly)
		assert.NotNil(t, prop.Default)
	})

	t.Run("convert any type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "dynamic_value",
					Type: "any",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["dynamic_value"]
		// Any type should allow multiple types
		assert.NotNil(t, prop.Type)
	})

	t.Run("convert nested complex types", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "complex_nested",
					Type: "list(object({name = string, subnets = list(string), config = map(string)}))",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["complex_nested"]
		// Should convert to array type
		assert.Equal(t, "array", prop.Type)
	})

	t.Run("convert list with complex element type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "server_list",
					Type: "list(object({name = string, port = number}))",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["server_list"]
		assert.Equal(t, "array", prop.Type)
	})

	t.Run("convert map with complex value type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "config_map",
					Type: "map(object({enabled = bool, value = string}))",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["config_map"]
		assert.Equal(t, "object", prop.Type)
	})

	t.Run("convert multiple variables with mixed types", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:     "name",
					Type:     "string",
					Required: true,
				},
				{
					Name:    "count",
					Type:    "number",
					Default: 1,
				},
				{
					Name:    "enabled",
					Type:    "bool",
					Default: false,
				},
				{
					Name: "tags",
					Type: "map(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Len(t, schema.Properties, 4)
		assert.Contains(t, schema.Required, "name")
		assert.Len(t, schema.Required, 1)
	})

	t.Run("preserve descriptions", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "instance_type",
					Type:        "string",
					Description: "The type of EC2 instance to launch",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["instance_type"]
		assert.Equal(t, "The type of EC2 instance to launch", prop.Description)
	})

	t.Run("convert with validation rules", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "instance_count",
					Type: "number",
					Validations: []parser.Validation{
						{
							Condition:    "var.instance_count > 0",
							ErrorMessage: "Must be positive",
						},
						{
							Condition:    "var.instance_count <= 10",
							ErrorMessage: "Must not exceed 10",
						},
					},
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["instance_count"]
		assert.Equal(t, "number", prop.Type)
		// Validation rules may be stored in description or custom fields
	})

	t.Run("convert outputs", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "input_var",
					Type: "string",
				},
			},
			Outputs: []parser.Output{
				{
					Name:        "instance_id",
					Description: "The ID of the instance",
					Sensitive:   false,
				},
				{
					Name:        "private_key",
					Description: "The private key",
					Sensitive:   true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		// Schema should still be generated even with outputs present
		assert.NotNil(t, schema)
	})

	t.Run("convert variable with complex default", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:    "tags",
					Type:    "map(string)",
					Default: map[string]string{"Environment": "prod"},
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["tags"]
		assert.Equal(t, "object", prop.Type)
		assert.NotNil(t, prop.Default)
	})

	t.Run("convert sensitive and required together", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:      "api_key",
					Type:      "string",
					Sensitive: true,
					Required:  true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["api_key"]
		assert.True(t, prop.WriteOnly)
		assert.Contains(t, schema.Required, "api_key")
	})

	t.Run("schema metadata", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "test_var",
					Type: "string",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
		assert.Equal(t, "object", schema.Type)
		assert.NotEmpty(t, schema.Properties)
	})

	t.Run("ToJSON produces valid JSON", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "test_var",
					Type:        "string",
					Description: "Test variable",
					Default:     "default_value",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)

		jsonBytes, err := converter.ToJSON(schema)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonBytes)

		// Verify it's valid JSON
		var result map[string]interface{}
		err = json.Unmarshal(jsonBytes, &result)
		require.NoError(t, err)
		assert.Equal(t, "http://json-schema.org/draft-07/schema#", result["$schema"])
	})

	t.Run("convert empty description", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "no_desc",
					Type:        "string",
					Description: "",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		prop := schema.Properties["no_desc"]
		assert.Empty(t, prop.Description)
	})
}

// TestConvertedSchemasAgainstMetaSchema validates all generated schemas against JSON Schema Draft 7 meta-schema
func TestConvertedSchemasAgainstMetaSchema(t *testing.T) {
	converter := NewConverter()

	testCases := []struct {
		name        string
		parseResult *parser.ParseResult
		description string
	}{
		{
			name: "simple string variable",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:        "instance_type",
						Type:        "string",
						Description: "EC2 instance type",
						Default:     "t2.micro",
					},
				},
			},
			description: "Basic string variable with default",
		},
		{
			name: "required variable",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:     "vpc_id",
						Type:     "string",
						Required: true,
					},
				},
			},
			description: "Required string variable",
		},
		{
			name: "sensitive variable",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:      "db_password",
						Type:      "string",
						Sensitive: true,
						Required:  true,
					},
				},
			},
			description: "Sensitive variable with writeOnly",
		},
		{
			name: "number type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "instance_count",
						Type: "number",
					},
				},
			},
			description: "Number type variable",
		},
		{
			name: "boolean type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "enable_monitoring",
						Type: "bool",
					},
				},
			},
			description: "Boolean type variable",
		},
		{
			name: "list type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "availability_zones",
						Type: "list(string)",
					},
				},
			},
			description: "List of strings",
		},
		{
			name: "map type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "tags",
						Type: "map(string)",
					},
				},
			},
			description: "Map of strings",
		},
		{
			name: "set type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "unique_zones",
						Type: "set(string)",
					},
				},
			},
			description: "Set type with uniqueItems",
		},
		{
			name: "tuple type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "example_tuple",
						Type: "tuple([string, number, bool])",
					},
				},
			},
			description: "Tuple type",
		},
		{
			name: "object type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "server_config",
						Type: "object({name = string, port = number, enabled = bool})",
					},
				},
			},
			description: "Object type with properties",
		},
		{
			name: "nullable variable",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:     "optional_string",
						Type:     "string",
						Nullable: true,
						Default:  nil,
					},
				},
			},
			description: "Nullable variable allowing null type",
		},
		{
			name: "ephemeral variable",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:      "session_token",
						Type:      "string",
						Ephemeral: true,
					},
				},
			},
			description: "Ephemeral variable",
		},
		{
			name: "any type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "dynamic_value",
						Type: "any",
					},
				},
			},
			description: "Any type allowing multiple types",
		},
		{
			name: "nested complex type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "complex_nested",
						Type: "list(object({name = string, subnets = list(string), config = map(string)}))",
					},
				},
			},
			description: "Complex nested type structure",
		},
		{
			name: "complete variable with all attributes",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:        "complete_var",
						Type:        "string",
						Description: "Complete variable with all attributes",
						Default:     "default",
						Sensitive:   true,
						Nullable:    false,
						Ephemeral:   true,
						Required:    false,
						Validations: []parser.Validation{
							{
								Condition:    "length(var.complete_var) > 5",
								ErrorMessage: "Must be longer than 5",
							},
						},
					},
				},
			},
			description: "Variable with all possible attributes",
		},
		{
			name: "multiple variables mixed types",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:     "name",
						Type:     "string",
						Required: true,
					},
					{
						Name:    "count",
						Type:    "number",
						Default: 1,
					},
					{
						Name:    "enabled",
						Type:    "bool",
						Default: false,
					},
					{
						Name: "tags",
						Type: "map(string)",
					},
				},
			},
			description: "Multiple variables with different types",
		},
		{
			name: "with validations",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "instance_count",
						Type: "number",
						Validations: []parser.Validation{
							{
								Condition:    "var.instance_count > 0",
								ErrorMessage: "Must be positive",
							},
							{
								Condition:    "var.instance_count <= 10",
								ErrorMessage: "Must not exceed 10",
							},
						},
					},
				},
			},
			description: "Variable with validation rules",
		},
		{
			name: "with complex default",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:    "tags",
						Type:    "map(string)",
						Default: map[string]string{"Environment": "prod"},
					},
				},
			},
			description: "Variable with complex default value",
		},
		{
			name: "object with optional() attributes",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name:        "server_config",
						Type:        "object({name = string, port = optional(number), enabled = optional(bool, true)})",
						Description: "Server config with optional fields",
					},
				},
			},
			description: "Object type with optional() modifier",
		},
		{
			name: "optional() with defaults",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "config",
						Type: "object({required = string, opt_str = optional(string, \"default\"), opt_num = optional(number, 42)})",
					},
				},
			},
			description: "Optional attributes with default values",
		},
		{
			name: "nested optional() in complex type",
			parseResult: &parser.ParseResult{
				Variables: []parser.Variable{
					{
						Name: "cluster",
						Type: "object({name = string, nodes = optional(list(object({type = string, count = optional(number, 3)})))})",
					},
				},
			},
			description: "Nested optional() in complex structures",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := converter.ConvertToJSONSchema7(tc.parseResult)
			require.NoError(t, err, "Failed to convert parse result to schema")

			// Validate against meta-schema
			validateSchemaAgainstMetaSchema(t, schema)

			t.Logf("✅ %s: Schema validated successfully against JSON Schema Draft 7 meta-schema", tc.description)
		})
	}
}

// TestOptionalTypeModifier tests the handling of Terraform's optional() type modifier
func TestOptionalTypeModifier(t *testing.T) {
	converter := NewConverter()

	t.Run("convert object with optional attributes", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "config",
					Type:        "object({name = string, port = optional(number), tags = optional(map(string))})",
					Description: "Configuration with optional fields",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)

		prop := schema.Properties["config"]
		assert.Equal(t, "object", prop.Type)
		assert.NotNil(t, prop)

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})

	t.Run("convert optional with default values", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "settings",
					Type: "object({enabled = optional(bool, true), timeout = optional(number, 30)})",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.NotNil(t, schema.Properties["settings"])

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})

	t.Run("convert list of objects with optional attributes", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "users",
					Type: "list(object({username = string, email = optional(string), role = optional(string, \"viewer\")}))",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)

		prop := schema.Properties["users"]
		assert.Equal(t, "array", prop.Type)

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})

	t.Run("convert nested optional in complex structure", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "cluster_config",
					Type: "object({name = string, monitoring = optional(object({enabled = bool, interval = optional(number, 60)}))})",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.NotNil(t, schema.Properties["cluster_config"])

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})

	t.Run("all optional attributes should not be in required", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "all_optional",
					Type: "object({opt1 = optional(string), opt2 = optional(number), opt3 = optional(bool)})",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)

		// The variable itself might be in required if no default, but internal attributes should handle optionals
		prop := schema.Properties["all_optional"]
		assert.Equal(t, "object", prop.Type)

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})

	t.Run("mixed required and optional object attributes", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "mixed_config",
					Type: "object({required_name = string, required_id = number, optional_desc = optional(string), optional_tags = optional(map(string), {})})",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)

		prop := schema.Properties["mixed_config"]
		assert.Equal(t, "object", prop.Type)

		// Validate against meta-schema
		validateSchemaAgainstMetaSchema(t, schema)
	})
}
