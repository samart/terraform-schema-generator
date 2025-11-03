package generator

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTerraformConfig = `
variable "test_var" {
  type        = string
  description = "Test variable"
}

variable "optional_var" {
  type    = number
  default = 42
}
`

func TestGenerator_FluentAPI(t *testing.T) {
	t.Run("simple fluent chain", func(t *testing.T) {
		schemaJSON, err := New().
			FromString("test.tf", testTerraformConfig).
			Parse().
			Convert().
			Validate().
			JSON()

		require.NoError(t, err)
		assert.NotEmpty(t, schemaJSON)
		assert.Contains(t, string(schemaJSON), `"$schema"`)
		assert.Contains(t, string(schemaJSON), `"test_var"`)
	})

	t.Run("with meta-schema validation", func(t *testing.T) {
		err := New().
			FromString("test.tf", testTerraformConfig).
			Parse().
			Convert().
			Validate().
			ValidateAgainstMetaSchema().
			Error()

		assert.NoError(t, err)
	})

	t.Run("access intermediate results", func(t *testing.T) {
		gen := New().
			FromString("test.tf", testTerraformConfig).
			Parse()

		parseResult, err := gen.ParseResult()
		require.NoError(t, err)
		assert.Len(t, parseResult.Variables, 2)

		schema, err := gen.Convert().Schema()
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Type)
		assert.Len(t, schema.Properties, 2)
	})

	t.Run("multiple files", func(t *testing.T) {
		vars := `variable "var1" { type = string }`
		outputs := `output "out1" { value = "test" }`

		schemaJSON, err := New().
			FromString("variables.tf", vars).
			FromString("outputs.tf", outputs).
			Parse().
			Convert().
			JSON()

		require.NoError(t, err)
		assert.Contains(t, string(schemaJSON), `"var1"`)
	})

	t.Run("error handling - parse before convert", func(t *testing.T) {
		gen := New()
		_, err := gen.Convert().JSON()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "call Parse() first")
	})

	t.Run("error handling - empty variables", func(t *testing.T) {
		emptyTF := `# No variables`

		err := New().
			FromString("empty.tf", emptyTF).
			Parse().
			Convert().
			Error()

		// Converting empty variables should error
		assert.Error(t, err)
	})
}

func TestGenerator_FromFile(t *testing.T) {
	t.Run("read from file", func(t *testing.T) {
		// Create temporary file
		tmpFile, err := os.CreateTemp("", "test-*.tf")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(testTerraformConfig)
		require.NoError(t, err)
		tmpFile.Close()

		schemaJSON, err := New().
			FromFile(tmpFile.Name()).
			Parse().
			Convert().
			JSON()

		require.NoError(t, err)
		assert.NotEmpty(t, schemaJSON)
	})

	t.Run("error on nonexistent file", func(t *testing.T) {
		err := New().
			FromFile("/nonexistent/file.tf").
			Error()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to open file")
	})
}

func TestGenerator_ToFile(t *testing.T) {
	t.Run("write to file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "schema-*.json")
		require.NoError(t, err)
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())

		err = New().
			FromString("test.tf", testTerraformConfig).
			Parse().
			Convert().
			ToFile(tmpFile.Name()).
			Error()

		require.NoError(t, err)

		// Verify file was written
		content, err := os.ReadFile(tmpFile.Name())
		require.NoError(t, err)
		assert.Contains(t, string(content), `"$schema"`)
	})
}

func TestGenerator_FromDirectory(t *testing.T) {
	t.Run("read from directory", func(t *testing.T) {
		// Create temporary directory with .tf files
		tmpDir, err := os.MkdirTemp("", "terraform-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		// Write test files
		err = os.WriteFile(tmpDir+"/variables.tf", []byte(`variable "var1" { type = string }`), 0644)
		require.NoError(t, err)

		err = os.WriteFile(tmpDir+"/outputs.tf", []byte(`output "out1" { value = "test" }`), 0644)
		require.NoError(t, err)

		// Add a non-.tf file that should be ignored
		err = os.WriteFile(tmpDir+"/README.md", []byte("# Test"), 0644)
		require.NoError(t, err)

		schemaJSON, err := New().
			FromDirectory(tmpDir).
			Parse().
			Convert().
			JSON()

		require.NoError(t, err)
		assert.Contains(t, string(schemaJSON), `"var1"`)
	})

	t.Run("error on nonexistent directory", func(t *testing.T) {
		err := New().
			FromDirectory("/nonexistent/directory").
			Error()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read directory")
	})
}

func TestGenerator_QuickHelpers(t *testing.T) {
	t.Run("FromStringQuick", func(t *testing.T) {
		schemaJSON, err := FromStringQuick(testTerraformConfig)

		require.NoError(t, err)
		assert.NotEmpty(t, schemaJSON)
		assert.Contains(t, string(schemaJSON), `"test_var"`)
	})

	t.Run("FromFileQuick", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "test-*.tf")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(testTerraformConfig)
		require.NoError(t, err)
		tmpFile.Close()

		schemaJSON, err := FromFileQuick(tmpFile.Name())

		require.NoError(t, err)
		assert.NotEmpty(t, schemaJSON)
	})

	t.Run("FromDirectoryQuick", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "terraform-*")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = os.WriteFile(tmpDir+"/main.tf", []byte(testTerraformConfig), 0644)
		require.NoError(t, err)

		schemaJSON, err := FromDirectoryQuick(tmpDir)

		require.NoError(t, err)
		assert.NotEmpty(t, schemaJSON)
	})
}

func TestGenerator_ErrorAccumulation(t *testing.T) {
	t.Run("errors accumulate", func(t *testing.T) {
		gen := New().
			FromFile("/nonexistent1.tf").
			FromFile("/nonexistent2.tf")

		errors := gen.Errors()
		assert.Len(t, errors, 2)
	})

	t.Run("first error returned", func(t *testing.T) {
		err := New().
			FromFile("/nonexistent1.tf").
			FromFile("/nonexistent2.tf").
			Error()

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent1.tf")
	})

	t.Run("errors prevent further execution", func(t *testing.T) {
		gen := New().
			FromString("test.tf", "invalid").
			Parse() // This will fail

		// These should not execute
		gen.Convert().Validate()

		errors := gen.Errors()
		// Should only have parse error, not conversion errors
		assert.Len(t, errors, 1)
	})
}

func TestStringReader(t *testing.T) {
	t.Run("read full content", func(t *testing.T) {
		sr := &stringReader{content: "hello world", pos: 0}

		buf := make([]byte, 100)
		n, err := sr.Read(buf)

		assert.NoError(t, err)
		assert.Equal(t, 11, n)
		assert.Equal(t, "hello world", string(buf[:n]))
	})

	t.Run("read in chunks", func(t *testing.T) {
		sr := &stringReader{content: "hello world", pos: 0}

		buf := make([]byte, 5)
		n, _ := sr.Read(buf)
		assert.Equal(t, "hello", string(buf[:n]))

		n, _ = sr.Read(buf)
		assert.Equal(t, " worl", string(buf[:n]))

		n, _ = sr.Read(buf)
		assert.Equal(t, "d", string(buf[:n]))
	})

	t.Run("EOF at end", func(t *testing.T) {
		sr := &stringReader{content: "test", pos: 4}

		buf := make([]byte, 10)
		n, err := sr.Read(buf)

		assert.Equal(t, 0, n)
		assert.Error(t, err)
	})
}

func TestGenerator_WithOptionalTypes(t *testing.T) {
	t.Run("handles optional() type modifier", func(t *testing.T) {
		tfWithOptional := `
variable "config" {
  type = object({
    name = string
    port = optional(number, 8080)
    enabled = optional(bool, true)
  })
}
`

		schemaJSON, err := New().
			FromString("test.tf", tfWithOptional).
			Parse().
			Convert().
			ValidateAgainstMetaSchema().
			JSON()

		require.NoError(t, err)
		assert.Contains(t, string(schemaJSON), `"config"`)
	})
}
