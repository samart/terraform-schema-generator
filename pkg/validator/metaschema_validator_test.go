package validator

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaSchemaValidator_ValidateAgainstMetaSchema(t *testing.T) {
	validator := NewMetaSchemaValidator()

	t.Run("valid simple schema", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("valid schema with multiple types", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
				"age": map[string]interface{}{
					"type": "number",
				},
				"active": map[string]interface{}{
					"type": "boolean",
				},
			},
			"required": []string{"name"},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("valid schema with array type", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "array",
			"items": map[string]interface{}{
				"type": "string",
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("valid schema with nested objects", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"address": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"street": map[string]interface{}{
							"type": "string",
						},
						"city": map[string]interface{}{
							"type": "string",
						},
					},
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("valid schema with constraints", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":      "string",
					"minLength": 1,
					"maxLength": 100,
				},
				"age": map[string]interface{}{
					"type":    "number",
					"minimum": 0,
					"maximum": 120,
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("valid schema with writeOnly", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"password": map[string]interface{}{
					"type":      "string",
					"writeOnly": true,
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})

	t.Run("invalid schema - wrong type value", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "invalid_type", // Invalid type
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not valid against JSON Schema Draft 7 meta-schema")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		invalidJSON := []byte(`{"type": "object", invalid}`)

		err := validator.ValidateAgainstMetaSchema(invalidJSON)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse schema JSON")
	})

	t.Run("schema with additional properties", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
			"additionalProperties": false,
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err)
	})
}

func TestMetaSchemaValidator_ValidateAgainstMetaSchemaWithDetails(t *testing.T) {
	validator := NewMetaSchemaValidator()

	t.Run("valid schema returns no errors", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type": "string",
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		result, err := validator.ValidateAgainstMetaSchemaWithDetails(schemaBytes)
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid schema returns errors", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "invalid_type",
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		result, err := validator.ValidateAgainstMetaSchemaWithDetails(schemaBytes)
		require.NoError(t, err)
		assert.False(t, result.Valid)
		assert.NotEmpty(t, result.Errors)
	})
}

func TestMetaSchemaValidator_RealWorldSchemas(t *testing.T) {
	validator := NewMetaSchemaValidator()

	t.Run("comprehensive Terraform schema", func(t *testing.T) {
		// This represents a realistic Terraform variables schema
		schema := map[string]interface{}{
			"$schema":     "http://json-schema.org/draft-07/schema#",
			"title":       "Terraform Variables Schema",
			"description": "Generated JSON Schema from Terraform variable definitions",
			"type":        "object",
			"properties": map[string]interface{}{
				"instance_type": map[string]interface{}{
					"type":        "string",
					"description": "EC2 instance type",
					"default":     "t2.micro",
				},
				"instance_count": map[string]interface{}{
					"type":        "number",
					"description": "Number of instances",
					"minimum":     float64(1),
					"maximum":     float64(10),
				},
				"enable_monitoring": map[string]interface{}{
					"type":        "boolean",
					"description": "Enable CloudWatch monitoring",
					"default":     false,
				},
				"tags": map[string]interface{}{
					"type":        "object",
					"description": "Resource tags",
					"default":     map[string]interface{}{},
				},
				"subnet_ids": map[string]interface{}{
					"type":        "array",
					"description": "List of subnet IDs",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"db_password": map[string]interface{}{
					"type":        "string",
					"description": "Database password",
					"writeOnly":   true,
				},
			},
			"required": []string{"subnet_ids"},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err, "Comprehensive Terraform schema should be valid")
	})

	t.Run("schema with complex nested structures", func(t *testing.T) {
		schema := map[string]interface{}{
			"$schema": "http://json-schema.org/draft-07/schema#",
			"type":    "object",
			"properties": map[string]interface{}{
				"cluster_configuration": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"execute_command_configuration": map[string]interface{}{
							"type": "object",
							"properties": map[string]interface{}{
								"kms_key_id": map[string]interface{}{
									"type": "string",
								},
								"logging": map[string]interface{}{
									"type": "string",
								},
							},
						},
					},
				},
			},
		}

		schemaBytes, err := json.Marshal(schema)
		require.NoError(t, err)

		err = validator.ValidateAgainstMetaSchema(schemaBytes)
		assert.NoError(t, err, "Nested structure schema should be valid")
	})
}
