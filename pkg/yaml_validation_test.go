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

// TestYAMLExamplesValidation validates all YAML example files against their respective schemas
func TestYAMLExamplesValidation(t *testing.T) {
	testCases := []struct {
		name          string
		schemaPath    string
		yamlPath      string
		expectedProps int // Expected number of top-level properties
		mustHaveProps []string
	}{
		{
			name:          "DynamoDB Basic Example",
			schemaPath:    filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json"),
			yamlPath:      filepath.Join("..", "testdata", "examples", "dynamodb-basic", "inputs.yaml"),
			expectedProps: 9,
			mustHaveProps: []string{"name", "hash_key", "range_key", "attributes", "tags"},
		},
		{
			name:          "DynamoDB Autoscaling Example",
			schemaPath:    filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json"),
			yamlPath:      filepath.Join("..", "testdata", "examples", "dynamodb-autoscaling", "inputs.yaml"),
			expectedProps: 14,
			mustHaveProps: []string{"name", "billing_mode", "autoscaling_enabled", "autoscaling_read", "autoscaling_write"},
		},
		{
			name:          "ECS Simple Example",
			schemaPath:    filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json"),
			yamlPath:      filepath.Join("..", "testdata", "examples", "ecs-simple", "inputs.yaml"),
			expectedProps: 9,
			mustHaveProps: []string{"create", "cluster_name", "region", "tags"},
		},
		{
			name:          "ECS Fargate Complex Example",
			schemaPath:    filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json"),
			yamlPath:      filepath.Join("..", "testdata", "examples", "ecs-fargate", "inputs.yaml"),
			expectedProps: 13,
			mustHaveProps: []string{"cluster_name", "services", "default_capacity_provider_strategy"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if schema exists
			_, err := os.Stat(tc.schemaPath)
			if os.IsNotExist(err) {
				t.Skipf("Schema file not found: %s", tc.schemaPath)
			}

			// Check if YAML exists
			_, err = os.Stat(tc.yamlPath)
			if os.IsNotExist(err) {
				t.Skipf("YAML file not found: %s", tc.yamlPath)
			}

			// Read and parse YAML
			yamlData, err := os.ReadFile(tc.yamlPath)
			require.NoError(t, err, "Failed to read YAML file")

			var inputData map[string]interface{}
			err = yaml.Unmarshal(yamlData, &inputData)
			require.NoError(t, err, "Failed to parse YAML")

			// Verify expected properties count
			assert.Equal(t, tc.expectedProps, len(inputData),
				"Expected %d properties, got %d", tc.expectedProps, len(inputData))

			// Verify required properties exist
			for _, prop := range tc.mustHaveProps {
				assert.Contains(t, inputData, prop,
					"Expected property '%s' to exist", prop)
			}

			// Convert to JSON for validation
			inputJSON, err := json.Marshal(inputData)
			require.NoError(t, err, "Failed to convert to JSON")

			// Validate against schema
			schemaLoader := gojsonschema.NewReferenceLoader("file://" + tc.schemaPath)
			documentLoader := gojsonschema.NewBytesLoader(inputJSON)

			result, err := gojsonschema.Validate(schemaLoader, documentLoader)
			require.NoError(t, err, "Validation error")

			// Check validation result
			if !result.Valid() {
				t.Errorf("❌ Validation failed for %s:", tc.name)
				for _, validationErr := range result.Errors() {
					t.Errorf("  - %s", validationErr.String())
				}
				t.FailNow()
			}

			t.Logf("✅ %s validated successfully (%d properties)", tc.name, len(inputData))
		})
	}
}

// TestYAMLExampleStructure validates the structure of YAML examples
func TestYAMLExampleStructure(t *testing.T) {
	t.Run("DynamoDB Basic - Structure", func(t *testing.T) {
		yamlPath := filepath.Join("..", "testdata", "examples", "dynamodb-basic", "inputs.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			t.Skip("YAML file not found")
		}

		yamlData, err := os.ReadFile(yamlPath)
		require.NoError(t, err)

		var data map[string]interface{}
		err = yaml.Unmarshal(yamlData, &data)
		require.NoError(t, err)

		// Verify specific structure
		assert.Equal(t, "my-table-test", data["name"])
		assert.Equal(t, "id", data["hash_key"])
		assert.Equal(t, "title", data["range_key"])

		// Verify attributes is an array
		attributes, ok := data["attributes"].([]interface{})
		require.True(t, ok, "attributes should be an array")
		assert.Len(t, attributes, 3, "Should have 3 attributes")

		// Verify tags is a map
		tags, ok := data["tags"].(map[string]interface{})
		require.True(t, ok, "tags should be a map")
		assert.Contains(t, tags, "Terraform")
	})

	t.Run("ECS Fargate - Structure", func(t *testing.T) {
		yamlPath := filepath.Join("..", "testdata", "examples", "ecs-fargate", "inputs.yaml")
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			t.Skip("YAML file not found")
		}

		yamlData, err := os.ReadFile(yamlPath)
		require.NoError(t, err)

		var data map[string]interface{}
		err = yaml.Unmarshal(yamlData, &data)
		require.NoError(t, err)

		// Verify cluster name
		assert.Equal(t, "ecs-fargate-example", data["cluster_name"])

		// Verify services is a map
		services, ok := data["services"].(map[string]interface{})
		require.True(t, ok, "services should be a map")
		assert.Contains(t, services, "frontend", "Should have frontend service")

		// Verify default_capacity_provider_strategy
		strategy, ok := data["default_capacity_provider_strategy"].(map[string]interface{})
		require.True(t, ok, "default_capacity_provider_strategy should be a map")
		assert.Contains(t, strategy, "FARGATE")
		assert.Contains(t, strategy, "FARGATE_SPOT")
	})
}

// TestAllYAMLExamplesExist verifies all expected YAML example files exist
func TestAllYAMLExamplesExist(t *testing.T) {
	expectedFiles := []struct {
		path        string
		description string
	}{
		{
			path:        filepath.Join("..", "testdata", "examples", "dynamodb-basic", "inputs.yaml"),
			description: "DynamoDB Basic Example",
		},
		{
			path:        filepath.Join("..", "testdata", "examples", "dynamodb-autoscaling", "inputs.yaml"),
			description: "DynamoDB Autoscaling Example",
		},
		{
			path:        filepath.Join("..", "testdata", "examples", "ecs-simple", "inputs.yaml"),
			description: "ECS Simple Example",
		},
		{
			path:        filepath.Join("..", "testdata", "examples", "ecs-fargate", "inputs.yaml"),
			description: "ECS Fargate Example",
		},
	}

	for _, file := range expectedFiles {
		t.Run(file.description, func(t *testing.T) {
			_, err := os.Stat(file.path)
			assert.NoError(t, err, "File should exist: %s", file.path)

			if err == nil {
				// Verify it's valid YAML
				yamlData, err := os.ReadFile(file.path)
				require.NoError(t, err)

				var data interface{}
				err = yaml.Unmarshal(yamlData, &data)
				assert.NoError(t, err, "Should be valid YAML")
			}
		})
	}
}

// TestYAMLExamplesAgainstSchemaDetailed provides detailed validation with specific assertions
func TestYAMLExamplesAgainstSchemaDetailed(t *testing.T) {
	t.Run("Validate DynamoDB Autoscaling Complex Config", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")
		yamlPath := filepath.Join("..", "testdata", "examples", "dynamodb-autoscaling", "inputs.yaml")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema not found")
		}
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			t.Skip("YAML not found")
		}

		// Read YAML
		yamlData, err := os.ReadFile(yamlPath)
		require.NoError(t, err)

		var data map[string]interface{}
		err = yaml.Unmarshal(yamlData, &data)
		require.NoError(t, err)

		// Verify complex nested structures
		assert.Equal(t, "PROVISIONED", data["billing_mode"])
		assert.Equal(t, true, data["autoscaling_enabled"])

		// Verify autoscaling_read structure
		autoscalingRead, ok := data["autoscaling_read"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, autoscalingRead, "max_capacity")
		assert.Contains(t, autoscalingRead, "target_value")

		// Validate against schema
		inputJSON, _ := json.Marshal(data)
		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)
		assert.True(t, result.Valid(), "Should validate successfully")
	})

	t.Run("Validate ECS Fargate Complex Services", func(t *testing.T) {
		schemaPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json")
		yamlPath := filepath.Join("..", "testdata", "examples", "ecs-fargate", "inputs.yaml")

		if _, err := os.Stat(schemaPath); os.IsNotExist(err) {
			t.Skip("Schema not found")
		}
		if _, err := os.Stat(yamlPath); os.IsNotExist(err) {
			t.Skip("YAML not found")
		}

		// Read YAML
		yamlData, err := os.ReadFile(yamlPath)
		require.NoError(t, err)

		var data map[string]interface{}
		err = yaml.Unmarshal(yamlData, &data)
		require.NoError(t, err)

		// Verify services structure
		services, ok := data["services"].(map[string]interface{})
		require.True(t, ok, "services should be a map")

		frontend, ok := services["frontend"].(map[string]interface{})
		require.True(t, ok, "frontend service should exist")

		// Verify container definitions
		containerDefs, ok := frontend["container_definitions"].(map[string]interface{})
		require.True(t, ok, "container_definitions should be a map")
		assert.Contains(t, containerDefs, "fluent-bit")
		assert.Contains(t, containerDefs, "ecsdemo-frontend")

		// Validate against schema
		inputJSON, _ := json.Marshal(data)
		schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
		documentLoader := gojsonschema.NewBytesLoader(inputJSON)

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		require.NoError(t, err)

		if !result.Valid() {
			for _, validationErr := range result.Errors() {
				t.Logf("Validation error: %s", validationErr.String())
			}
		}
		assert.True(t, result.Valid(), "Should validate successfully")
	})
}
