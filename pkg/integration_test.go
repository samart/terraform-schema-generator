package pkg_test

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/samart/terraform-schema-generator/pkg/converter"
	"github.com/samart/terraform-schema-generator/pkg/parser"
	"github.com/samart/terraform-schema-generator/pkg/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

// validateYAMLAgainstSchema validates a YAML file against a JSON schema
func validateYAMLAgainstSchema(t *testing.T, yamlPath, schemaPath string) bool {
	// Read YAML file
	yamlData, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Logf("Warning: Could not read YAML file %s: %v", yamlPath, err)
		return false
	}

	// Parse YAML to map
	var inputData map[string]interface{}
	err = yaml.Unmarshal(yamlData, &inputData)
	if err != nil {
		t.Logf("Warning: Could not parse YAML file %s: %v", yamlPath, err)
		return false
	}

	// Convert to JSON
	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		t.Logf("Warning: Could not convert YAML to JSON: %v", err)
		return false
	}

	// Load schema and validate
	schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
	documentLoader := gojsonschema.NewBytesLoader(inputJSON)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		t.Logf("Warning: Validation error: %v", err)
		return false
	}

	if !result.Valid() {
		t.Logf("❌ YAML validation failed for %s:", yamlPath)
		for _, err := range result.Errors() {
			t.Logf("  - %s", err.String())
		}
		return false
	}

	t.Logf("✅ YAML validated: %s (%d properties)", filepath.Base(yamlPath), len(inputData))
	return true
}

// validateExampleInputs validates all example YAML files against a schema
func validateExampleInputs(t *testing.T, schemaPath, examplesPattern string) int {
	validCount := 0

	// Find all YAML files matching the pattern
	matches, err := filepath.Glob(examplesPattern)
	if err != nil || len(matches) == 0 {
		t.Logf("No example YAML files found matching: %s", examplesPattern)
		return 0
	}

	for _, yamlPath := range matches {
		if validateYAMLAgainstSchema(t, yamlPath, schemaPath) {
			validCount++
		}
	}

	return validCount
}

// TestIntegrationTerraformAWSECS tests parsing and converting the terraform-aws-ecs module variables
func TestIntegrationTerraformAWSECS(t *testing.T) {
	testdataPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "variables.tf")

	// Check if file exists
	_, err := os.Stat(testdataPath)
	if os.IsNotExist(err) {
		t.Skipf("Testdata file not found: %s", testdataPath)
	}

	// Read the variables file
	content, err := os.ReadFile(testdataPath)
	require.NoError(t, err, "Failed to read testdata file")
	require.NotEmpty(t, content, "Testdata file is empty")

	// Parse the Terraform variables
	p := parser.NewParser()
	files := map[string]io.Reader{
		"variables.tf": strings.NewReader(string(content)),
	}

	result, err := p.ParseFiles(files)
	require.NoError(t, err, "Failed to parse Terraform file")
	require.NotNil(t, result, "Parse result should not be nil")

	// Verify we found variables
	assert.Greater(t, len(result.Variables), 0, "Should have parsed at least one variable")

	// Log the number of variables found
	t.Logf("Parsed %d variables from terraform-aws-ecs", len(result.Variables))

	// Verify some expected variables exist
	variableNames := make(map[string]bool)
	for _, v := range result.Variables {
		variableNames[v.Name] = true
	}

	expectedVars := []string{
		"create",
		"region",
		"tags",
		"cluster_name",
		"cluster_configuration",
		"services",
	}

	for _, expectedVar := range expectedVars {
		assert.True(t, variableNames[expectedVar], "Expected variable '%s' to be present", expectedVar)
	}

	// Convert to JSON Schema
	c := converter.NewConverter()
	schema, err := c.ConvertToJSONSchema7(result)
	require.NoError(t, err, "Failed to convert to JSON Schema")
	require.NotNil(t, schema, "Schema should not be nil")

	// Verify schema structure
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
	assert.Equal(t, "object", schema.Type)
	assert.Greater(t, len(schema.Properties), 0, "Schema should have properties")

	// Verify the schema can be marshaled to JSON
	schemaJSON, err := c.ToJSON(schema)
	require.NoError(t, err, "Failed to marshal schema to JSON")
	assert.NotEmpty(t, schemaJSON, "Schema JSON should not be empty")

	// Log schema size
	t.Logf("Generated JSON Schema size: %d bytes", len(schemaJSON))

	// Validate against JSON Schema Draft 7 meta-schema
	metaValidator := validator.NewMetaSchemaValidator()
	err = metaValidator.ValidateAgainstMetaSchema(schemaJSON)
	require.NoError(t, err, "Generated schema should be valid against JSON Schema Draft 7 meta-schema")
	t.Logf("✅ Schema validated successfully against JSON Schema Draft 7 meta-schema")

	// Optionally write schema to file for inspection
	outputPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json")
	err = os.WriteFile(outputPath, schemaJSON, 0644)
	if err != nil {
		t.Logf("Warning: Could not write schema to %s: %v", outputPath, err)
	} else {
		t.Logf("Schema written to: %s", outputPath)
	}

	// Validate example YAML inputs against the generated schema
	t.Log("\n📋 Validating example inputs against schema...")
	examplesPattern := filepath.Join("..", "testdata", "examples", "ecs-*/inputs.yaml")
	validatedCount := validateExampleInputs(t, outputPath, examplesPattern)
	if validatedCount > 0 {
		t.Logf("✅ Successfully validated %d ECS example input(s)", validatedCount)
	}
}

// TestIntegrationTerraformAWSDynamoDBTable tests parsing and converting the terraform-aws-dynamodb-table module variables
func TestIntegrationTerraformAWSDynamoDBTable(t *testing.T) {
	testdataPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "variables.tf")

	// Check if file exists
	_, err := os.Stat(testdataPath)
	if os.IsNotExist(err) {
		t.Skipf("Testdata file not found: %s", testdataPath)
	}

	// Read the variables file
	content, err := os.ReadFile(testdataPath)
	require.NoError(t, err, "Failed to read testdata file")
	require.NotEmpty(t, content, "Testdata file is empty")

	// Parse the Terraform variables
	p := parser.NewParser()
	files := map[string]io.Reader{
		"variables.tf": strings.NewReader(string(content)),
	}

	result, err := p.ParseFiles(files)
	require.NoError(t, err, "Failed to parse Terraform file")
	require.NotNil(t, result, "Parse result should not be nil")

	// Verify we found variables
	assert.Greater(t, len(result.Variables), 0, "Should have parsed at least one variable")

	// Log the number of variables found
	t.Logf("Parsed %d variables from terraform-aws-dynamodb-table", len(result.Variables))

	// Verify some expected variables exist
	variableNames := make(map[string]bool)
	for _, v := range result.Variables {
		variableNames[v.Name] = true
	}

	expectedVars := []string{
		"create_table",
		"name",
		"attributes",
		"hash_key",
		"range_key",
		"billing_mode",
		"tags",
	}

	for _, expectedVar := range expectedVars {
		assert.True(t, variableNames[expectedVar], "Expected variable '%s' to be present", expectedVar)
	}

	// Verify specific variable properties
	for _, v := range result.Variables {
		switch v.Name {
		case "create_table":
			assert.Equal(t, "bool", v.Type, "create_table should be bool type")
			assert.False(t, v.Required, "create_table should have a default")
		case "name":
			assert.Equal(t, "string", v.Type, "name should be string type")
		case "billing_mode":
			assert.Equal(t, "string", v.Type, "billing_mode should be string type")
			assert.False(t, v.Required, "billing_mode should have a default")
		case "tags":
			assert.Contains(t, v.Type, "map", "tags should be map type")
		case "attributes":
			assert.Contains(t, v.Type, "list", "attributes should be list type")
		}
	}

	// Convert to JSON Schema
	c := converter.NewConverter()
	schema, err := c.ConvertToJSONSchema7(result)
	require.NoError(t, err, "Failed to convert to JSON Schema")
	require.NotNil(t, schema, "Schema should not be nil")

	// Verify schema structure
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
	assert.Equal(t, "object", schema.Type)
	assert.Greater(t, len(schema.Properties), 0, "Schema should have properties")

	// Verify specific properties in schema
	assert.Contains(t, schema.Properties, "create_table")
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "billing_mode")

	// Verify property types
	assert.Equal(t, "boolean", schema.Properties["create_table"].Type)
	assert.Equal(t, "string", schema.Properties["name"].Type)
	assert.Equal(t, "string", schema.Properties["billing_mode"].Type)

	// Verify descriptions are preserved
	assert.NotEmpty(t, schema.Properties["create_table"].Description)
	assert.NotEmpty(t, schema.Properties["name"].Description)

	// Verify the schema can be marshaled to JSON
	schemaJSON, err := c.ToJSON(schema)
	require.NoError(t, err, "Failed to marshal schema to JSON")
	assert.NotEmpty(t, schemaJSON, "Schema JSON should not be empty")

	// Log schema size
	t.Logf("Generated JSON Schema size: %d bytes", len(schemaJSON))

	// Validate against JSON Schema Draft 7 meta-schema
	metaValidator := validator.NewMetaSchemaValidator()
	err = metaValidator.ValidateAgainstMetaSchema(schemaJSON)
	require.NoError(t, err, "Generated schema should be valid against JSON Schema Draft 7 meta-schema")
	t.Logf("✅ Schema validated successfully against JSON Schema Draft 7 meta-schema")

	// Optionally write schema to file for inspection
	outputPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")
	err = os.WriteFile(outputPath, schemaJSON, 0644)
	if err != nil {
		t.Logf("Warning: Could not write schema to %s: %v", outputPath, err)
	} else {
		t.Logf("Schema written to: %s", outputPath)
	}

	// Validate example YAML inputs against the generated schema
	t.Log("\n📋 Validating example inputs against schema...")
	examplesPattern := filepath.Join("..", "testdata", "examples", "dynamodb-*/inputs.yaml")
	validatedCount := validateExampleInputs(t, outputPath, examplesPattern)
	if validatedCount > 0 {
		t.Logf("✅ Successfully validated %d DynamoDB example input(s)", validatedCount)
	}
}

// TestIntegrationBothModules runs a comprehensive test on both real-world modules
func TestIntegrationBothModules(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		minVars  int
		mustHave []string
	}{
		{
			name:     "terraform-aws-ecs",
			path:     filepath.Join("..", "testdata", "terraform-aws-ecs", "variables.tf"),
			minVars:  10,
			mustHave: []string{"create", "region", "tags"},
		},
		{
			name:     "terraform-aws-dynamodb-table",
			path:     filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "variables.tf"),
			minVars:  20,
			mustHave: []string{"create_table", "name", "hash_key"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Check if file exists
			_, err := os.Stat(tc.path)
			if os.IsNotExist(err) {
				t.Skipf("Testdata file not found: %s", tc.path)
			}

			// Read and parse
			content, err := os.ReadFile(tc.path)
			require.NoError(t, err)

			p := parser.NewParser()
			files := map[string]io.Reader{
				"variables.tf": strings.NewReader(string(content)),
			}

			result, err := p.ParseFiles(files)
			require.NoError(t, err)

			// Verify minimum variable count
			assert.GreaterOrEqual(t, len(result.Variables), tc.minVars,
				"Should have at least %d variables", tc.minVars)

			// Verify required variables exist
			varNames := make(map[string]bool)
			for _, v := range result.Variables {
				varNames[v.Name] = true
			}

			for _, mustHaveVar := range tc.mustHave {
				assert.True(t, varNames[mustHaveVar],
					"Expected variable '%s' to be present", mustHaveVar)
			}

			// Convert to schema
			c := converter.NewConverter()
			schema, err := c.ConvertToJSONSchema7(result)
			require.NoError(t, err)

			// Validate schema
			assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
			assert.Equal(t, "object", schema.Type)
			assert.GreaterOrEqual(t, len(schema.Properties), tc.minVars)

			// Ensure JSON serialization works
			schemaJSON, err := c.ToJSON(schema)
			require.NoError(t, err)

			// Validate against meta-schema
			metaValidator := validator.NewMetaSchemaValidator()
			err = metaValidator.ValidateAgainstMetaSchema(schemaJSON)
			require.NoError(t, err, "Schema should be valid against JSON Schema Draft 7 meta-schema")

			t.Logf("%s: Parsed %d variables, generated schema with %d properties, ✅ meta-schema valid",
				tc.name, len(result.Variables), len(schema.Properties))
		})
	}
}

// TestEndToEndWithExamples tests the complete workflow: parse -> convert -> validate schema -> validate examples
func TestEndToEndWithExamples(t *testing.T) {
	t.Run("Complete DynamoDB Workflow", func(t *testing.T) {
		// Step 1: Parse Terraform variables
		t.Log("Step 1: Parsing Terraform variables...")
		testdataPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "variables.tf")
		content, err := os.ReadFile(testdataPath)
		require.NoError(t, err)

		p := parser.NewParser()
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(string(content)),
		}

		result, err := p.ParseFiles(files)
		require.NoError(t, err)
		t.Logf("  ✅ Parsed %d variables", len(result.Variables))

		// Step 2: Convert to JSON Schema
		t.Log("Step 2: Converting to JSON Schema Draft 7...")
		c := converter.NewConverter()
		schema, err := c.ConvertToJSONSchema7(result)
		require.NoError(t, err)
		t.Logf("  ✅ Generated schema with %d properties", len(schema.Properties))

		// Step 3: Validate against meta-schema
		t.Log("Step 3: Validating against JSON Schema Draft 7 meta-schema...")
		schemaJSON, err := c.ToJSON(schema)
		require.NoError(t, err)

		metaValidator := validator.NewMetaSchemaValidator()
		err = metaValidator.ValidateAgainstMetaSchema(schemaJSON)
		require.NoError(t, err)
		t.Log("  ✅ Schema is valid JSON Schema Draft 7")

		// Step 4: Write schema to temp file for validation
		outputPath := filepath.Join("..", "testdata", "terraform-aws-dynamodb-table", "schema.json")
		err = os.WriteFile(outputPath, schemaJSON, 0644)
		require.NoError(t, err)

		// Step 5: Validate example YAML inputs
		t.Log("Step 4: Validating example inputs against schema...")
		examplesPattern := filepath.Join("..", "testdata", "examples", "dynamodb-*/inputs.yaml")
		validatedCount := validateExampleInputs(t, outputPath, examplesPattern)
		require.Greater(t, validatedCount, 0, "Should validate at least one example")
		t.Logf("  ✅ Validated %d example input(s)", validatedCount)

		// Summary
		t.Log("\n✅ End-to-End Test Complete:")
		t.Logf("  • Parsed: %d variables", len(result.Variables))
		t.Logf("  • Schema: %d properties, %d bytes", len(schema.Properties), len(schemaJSON))
		t.Logf("  • Examples validated: %d", validatedCount)
	})

	t.Run("Complete ECS Workflow", func(t *testing.T) {
		// Step 1: Parse Terraform variables
		t.Log("Step 1: Parsing Terraform variables...")
		testdataPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "variables.tf")
		content, err := os.ReadFile(testdataPath)
		require.NoError(t, err)

		p := parser.NewParser()
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(string(content)),
		}

		result, err := p.ParseFiles(files)
		require.NoError(t, err)
		t.Logf("  ✅ Parsed %d variables", len(result.Variables))

		// Step 2: Convert to JSON Schema
		t.Log("Step 2: Converting to JSON Schema Draft 7...")
		c := converter.NewConverter()
		schema, err := c.ConvertToJSONSchema7(result)
		require.NoError(t, err)
		t.Logf("  ✅ Generated schema with %d properties", len(schema.Properties))

		// Step 3: Validate against meta-schema
		t.Log("Step 3: Validating against JSON Schema Draft 7 meta-schema...")
		schemaJSON, err := c.ToJSON(schema)
		require.NoError(t, err)

		metaValidator := validator.NewMetaSchemaValidator()
		err = metaValidator.ValidateAgainstMetaSchema(schemaJSON)
		require.NoError(t, err)
		t.Log("  ✅ Schema is valid JSON Schema Draft 7")

		// Step 4: Write schema to temp file for validation
		outputPath := filepath.Join("..", "testdata", "terraform-aws-ecs", "schema.json")
		err = os.WriteFile(outputPath, schemaJSON, 0644)
		require.NoError(t, err)

		// Step 5: Validate example YAML inputs
		t.Log("Step 4: Validating example inputs against schema...")
		examplesPattern := filepath.Join("..", "testdata", "examples", "ecs-*/inputs.yaml")
		validatedCount := validateExampleInputs(t, outputPath, examplesPattern)
		require.Greater(t, validatedCount, 0, "Should validate at least one example")
		t.Logf("  ✅ Validated %d example input(s)", validatedCount)

		// Summary
		t.Log("\n✅ End-to-End Test Complete:")
		t.Logf("  • Parsed: %d variables", len(result.Variables))
		t.Logf("  • Schema: %d properties, %d bytes", len(schema.Properties), len(schemaJSON))
		t.Logf("  • Examples validated: %d", validatedCount)
	})
}
