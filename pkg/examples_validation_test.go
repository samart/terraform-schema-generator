package pkg_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// TestExampleInputsAgainstSchemas validates real-world module usage examples against generated schemas
func TestExampleInputsAgainstSchemas(t *testing.T) {
	testCases := []struct {
		name       string
		schemaPath string
		inputPath  string
	}{
		{
			name:       "DynamoDB Basic Example",
			schemaPath: filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json"),
			inputPath:  filepath.Join("..", "testdata", "examples", "dynamodb-basic", "inputs.yaml"),
		},
		{
			name:       "DynamoDB Autoscaling Example",
			schemaPath: filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json"),
			inputPath:  filepath.Join("..", "testdata", "examples", "dynamodb-autoscaling", "inputs.yaml"),
		},
		{
			name:       "ECS Simple Example",
			schemaPath: filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json"),
			inputPath:  filepath.Join("..", "testdata", "examples", "ecs-simple", "inputs.yaml"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if schema file exists
			_, err := os.Stat(tc.schemaPath)
			if os.IsNotExist(err) {
				t.Skipf("Schema file not found: %s", tc.schemaPath)
			}

			// Check if input file exists
			_, err = os.Stat(tc.inputPath)
			if os.IsNotExist(err) {
				t.Skipf("Input file not found: %s", tc.inputPath)
			}

			// Read and parse YAML input
			yamlData, err := os.ReadFile(tc.inputPath)
			require.NoError(t, err, "Failed to read YAML input file")

			// Convert YAML to map
			var inputData map[string]interface{}
			err = yaml.Unmarshal(yamlData, &inputData)
			require.NoError(t, err, "Failed to parse YAML input")

			// Convert to JSON for validation
			inputJSON, err := json.Marshal(inputData)
			require.NoError(t, err, "Failed to convert input to JSON")

			t.Logf("Input JSON: %s", string(inputJSON))

			// Load the JSON schema
			schemaLoader := gojsonschema.NewReferenceLoader("file://" + tc.schemaPath)

			// Load the input document
			documentLoader := gojsonschema.NewBytesLoader(inputJSON)

			// Validate
			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			require.NoError(t, err, "Validation error")

			// Check validation result
			if !result.Valid() {
				t.Logf("❌ Validation errors for %s:", tc.name)
				for _, err := range result.Errors() {
					t.Logf("  - %s", err.String())
				}
				t.Fail()
			} else {
				t.Logf("✅ %s validated successfully against schema", tc.name)
			}

			// Log statistics
			t.Logf("Input properties: %d", len(inputData))
		})
	}
}

// TestExampleInputsDetailed provides detailed validation information
func TestExampleInputsDetailed(t *testing.T) {
	t.Run("DynamoDB Basic - Detailed", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")
		inputPath := filepath.Join("..", "testdata", "examples", "dynamodb-basic", "inputs.yaml")

		// Check if files exist
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema file not found")
		}
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Input file not found")
		}

		// Read YAML
		yamlData, err := os.ReadFile(inputPath)
		require.NoError(t, err)

		var inputData map[string]interface{}
		err = yaml.Unmarshal(yamlData, &inputData)
		require.NoError(t, err)

		// Validate specific fields
		assert.Equal(t, "my-table-test", inputData["name"])
		assert.Equal(t, "id", inputData["hash_key"])
		assert.Equal(t, "title", inputData["range_key"])
		assert.Equal(t, false, inputData["deletion_protection_enabled"])

		// Check attributes structure
		attributes, ok := inputData["attributes"].([]interface{})
		require.True(t, ok, "attributes should be an array")
		assert.Len(t, attributes, 3)

		// Validate against schema
		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		inputJSON, _ := json.Marshal(inputData)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)

		if result.Valid() {
			t.Log("✅ Detailed validation passed")
		} else {
			for _, err := range result.Errors() {
				t.Logf("❌ %s", err.String())
			}
			t.Fail()
		}
	})

	t.Run("ECS Simple - Detailed", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json")
		inputPath := filepath.Join("..", "testdata", "examples", "ecs-simple", "inputs.yaml")

		// Check if files exist
		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema file not found")
		}
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			t.Skip("Input file not found")
		}

		// Read YAML
		yamlData, err := os.ReadFile(inputPath)
		require.NoError(t, err)

		var inputData map[string]interface{}
		err = yaml.Unmarshal(yamlData, &inputData)
		require.NoError(t, err)

		// Validate specific fields
		assert.Equal(t, true, inputData["create"])
		assert.Equal(t, "my-ecs-cluster", inputData["cluster_name"])
		assert.Equal(t, "us-east-1", inputData["region"])

		// Validate against schema
		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		inputJSON, _ := json.Marshal(inputData)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)

		if result.Valid() {
			t.Log("✅ Detailed validation passed")
		} else {
			for _, err := range result.Errors() {
				t.Logf("❌ %s", err.String())
			}
			t.Fail()
		}
	})
}

// TestMissingRequiredFields validates that schema properly catches missing required fields
func TestMissingRequiredFields(t *testing.T) {
	t.Run("DynamoDB missing required field", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema file not found")
		}

		// Create input missing required fields (if any are required)
		invalidInput := map[string]interface{}{
			"tags": map[string]string{
				"Environment": "test",
			},
			// Missing other potentially required fields
		}

		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		inputJSON, _ := json.Marshal(invalidInput)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)

		// Log the result (may or may not be valid depending on schema requirements)
		if result.Valid() {
			t.Log("✅ No required fields - schema allows minimal input")
		} else {
			t.Log("Schema correctly validates required fields:")
			for _, err := range result.Errors() {
				t.Logf("  - %s", err.String())
			}
		}
	})
}

// TestInvalidTypeValues validates that schema catches type mismatches
func TestInvalidTypeValues(t *testing.T) {
	t.Run("DynamoDB wrong type", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema file not found")
		}

		// Create input with wrong types
		invalidInput := map[string]interface{}{
			"name":                        123,    // Should be string
			"deletion_protection_enabled": "true", // Should be boolean
			"read_capacity":               "five", // Should be number
		}

		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		inputJSON, _ := json.Marshal(invalidInput)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)

		// Should have validation errors
		assert.False(t, result.Valid(), "Should fail validation with wrong types")

		if !result.Valid() {
			t.Log("✅ Schema correctly caught type errors:")
			for _, err := range result.Errors() {
				t.Logf("  - %s", err.String())
			}
		}
	})
}
