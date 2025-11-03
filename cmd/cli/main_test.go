package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestCommand creates a fresh command instance for testing
func setupTestCommand() *cobra.Command {
	// Reset global flags
	inputDir = ""
	inputFile = ""
	outputFile = ""
	validate = true
	verbose = false

	// Create a new root command instance
	cmd := &cobra.Command{
		Use:          "terraform-schema-generator",
		Short:        "Generate JSON Schema from Terraform configurations",
		Version:      version,
		RunE:         runGenerate,
		SilenceUsage: true,
	}

	// Add flags
	cmd.Flags().StringVarP(&inputDir, "dir", "d", "", "Directory containing Terraform files")
	cmd.Flags().StringVarP(&inputFile, "file", "f", "", "Single Terraform file to process")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().BoolVar(&validate, "validate", true, "Validate generated schema against JSON Schema Draft 7")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	cmd.MarkFlagsMutuallyExclusive("dir", "file")

	return cmd
}

// executeCommand executes a command with the given args and returns stdout, stderr, and error
func executeCommand(cmd *cobra.Command, args ...string) (stdout, stderr string, err error) {
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	stdoutBuf := new(bytes.Buffer)
	stderrBuf := new(bytes.Buffer)

	// Create pipes for capturing output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	os.Stdout = wOut
	os.Stderr = wErr

	cmd.SetArgs(args)
	cmd.SetOut(wOut)
	cmd.SetErr(wErr)

	// Execute command
	cmdErr := cmd.Execute()

	// Close writers and read output
	wOut.Close()
	wErr.Close()

	stdoutBuf.ReadFrom(rOut)
	stderrBuf.ReadFrom(rErr)

	return stdoutBuf.String(), stderrBuf.String(), cmdErr
}

func TestCLI_Version(t *testing.T) {
	cmd := setupTestCommand()
	stdout, _, err := executeCommand(cmd, "--version")

	require.NoError(t, err)
	assert.Contains(t, stdout, "terraform-schema-generator version")
	assert.Contains(t, stdout, version)
}

func TestCLI_Help(t *testing.T) {
	cmd := setupTestCommand()
	stdout, _, err := executeCommand(cmd, "--help")

	require.NoError(t, err)
	assert.Contains(t, stdout, "Generate JSON Schema from Terraform configurations")
	assert.Contains(t, stdout, "Usage:")
	assert.Contains(t, stdout, "Flags:")
	assert.Contains(t, stdout, "--dir")
	assert.Contains(t, stdout, "--file")
	assert.Contains(t, stdout, "--output")
	assert.Contains(t, stdout, "--validate")
	assert.Contains(t, stdout, "--verbose")
}

func TestCLI_NoInputSpecified(t *testing.T) {
	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd)

	require.Error(t, err)
	assert.Contains(t, stderr, "either --dir or --file must be specified")
}

func TestCLI_BothInputsSpecified(t *testing.T) {
	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", "test.tf", "-d", "testdir")

	require.Error(t, err)
	// Cobra catches mutually exclusive flags
	assert.True(t,
		strings.Contains(stderr, "if any flags in the group") ||
		strings.Contains(stderr, "only one of"),
		"Expected error about mutually exclusive flags, got: %s", stderr)
}

func TestCLI_NonExistentFile(t *testing.T) {
	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", "/nonexistent/file.tf")

	require.Error(t, err)
	assert.Contains(t, stderr, "cannot access file")
}

func TestCLI_NonExistentDirectory(t *testing.T) {
	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-d", "/nonexistent/directory")

	require.Error(t, err)
	assert.Contains(t, stderr, "cannot access directory")
}

func TestCLI_FileIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tmpDir)

	require.Error(t, err)
	assert.Contains(t, stderr, "path is a directory, not a file")
}

func TestCLI_DirectoryIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tf")
	require.NoError(t, os.WriteFile(tmpFile, []byte("# test"), 0644))

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-d", tmpFile)

	require.Error(t, err)
	assert.Contains(t, stderr, "path is not a directory")
}

func TestCLI_GenerateFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")

	tfContent := `
variable "cluster_name" {
  type        = string
  description = "Name of the cluster"
}

variable "instance_count" {
  type    = number
  default = 3
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate=false")

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify JSON Schema output
	assert.Contains(t, stdout, `"$schema"`)
	assert.Contains(t, stdout, `"cluster_name"`)
	assert.Contains(t, stdout, `"instance_count"`)
	assert.Contains(t, stdout, `"Name of the cluster"`)
}

func TestCLI_GenerateFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple .tf files
	tfContent1 := `
variable "vpc_id" {
  type = string
}
`
	tfContent2 := `
variable "subnet_ids" {
  type = list(string)
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "vpc.tf"), []byte(tfContent1), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "network.tf"), []byte(tfContent2), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-d", tmpDir, "--validate=false")

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify both variables are in output
	assert.Contains(t, stdout, `"vpc_id"`)
	assert.Contains(t, stdout, `"subnet_ids"`)
}

func TestCLI_OutputToFile(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")
	outputFile := filepath.Join(tmpDir, "schema.json")

	tfContent := `
variable "test_var" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "-o", outputFile, "--validate=false")

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify file was created
	assert.FileExists(t, outputFile)

	// Verify file contents
	schemaContent, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(schemaContent), `"test_var"`)

	// Stdout should be empty when writing to file
	assert.Empty(t, stdout)
}

func TestCLI_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")

	tfContent := `
variable "test" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "-v", "--validate=false")

	require.NoError(t, err)

	// Verbose output goes to stderr
	assert.Contains(t, stderr, "Reading Terraform file")
	assert.Contains(t, stderr, "Parsing Terraform configuration")
	assert.Contains(t, stderr, "Found 1 variables")
	assert.Contains(t, stderr, "Converting to JSON Schema")

	// Schema goes to stdout
	assert.Contains(t, stdout, `"test"`)
}

func TestCLI_VerboseWithOutputFile(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")
	outputFile := filepath.Join(tmpDir, "schema.json")

	tfContent := `
variable "test" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tfFile, "-o", outputFile, "-v", "--validate=false")

	require.NoError(t, err)

	// Verbose output includes file write confirmation
	assert.Contains(t, stderr, "Writing schema to file")
	assert.Contains(t, stderr, "successfully written")
	assert.Contains(t, stderr, outputFile)
}

func TestCLI_ValidationEnabled(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")

	tfContent := `
variable "test" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate", "-v")

	require.NoError(t, err, "stderr: %s", stderr)

	// With verbose, should see validation message
	assert.Contains(t, stderr, "Validating against JSON Schema Draft 7")
	assert.Contains(t, stderr, "validation passed")

	// Schema should still be output
	assert.Contains(t, stdout, `"test"`)
}

func TestCLI_ValidationDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")

	tfContent := `
variable "test" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate=false", "-v")

	require.NoError(t, err)

	// Should NOT see validation message
	assert.NotContains(t, stderr, "Validating against JSON Schema Draft 7")
}

func TestCLI_EmptyTerraformFile(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "empty.tf")

	// Empty file
	require.NoError(t, os.WriteFile(tfFile, []byte(""), 0644))

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate=false")

	require.Error(t, err)
	assert.Contains(t, stderr, "no variables to convert")
}

func TestCLI_InvalidTerraformSyntax(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "invalid.tf")

	invalidContent := `
variable "bad" {
  this is not valid HCL syntax!!!
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(invalidContent), 0644))

	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tfFile)

	require.Error(t, err)
	// Should fail either at parsing or conversion stage
	assert.True(t,
		strings.Contains(stderr, "parsing failed") ||
		strings.Contains(stderr, "conversion failed") ||
		strings.Contains(stderr, "no variables"),
		"Expected error about parsing or conversion failure, got: %s", stderr)
}

func TestCLI_ShortFlags(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "variables.tf")
	outputFile := filepath.Join(tmpDir, "schema.json")

	tfContent := `
variable "test" {
  type = string
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	// Test short flags: -f, -o, -v
	cmd := setupTestCommand()
	_, stderr, err := executeCommand(cmd, "-f", tfFile, "-o", outputFile, "-v", "--validate=false")

	require.NoError(t, err)

	// Verify verbose output (from -v)
	assert.Contains(t, stderr, "Reading Terraform file")

	// Verify output file was created (from -o)
	assert.FileExists(t, outputFile)
}

func TestCLI_ComplexTypes(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "complex.tf")

	tfContent := `
variable "simple_string" {
  type        = string
  description = "A simple string"
}

variable "number_var" {
  type    = number
  default = 42
}

variable "bool_var" {
  type    = bool
  default = true
}

variable "list_var" {
  type    = list(string)
  default = ["a", "b", "c"]
}

variable "map_var" {
  type = map(string)
  default = {
    key1 = "value1"
    key2 = "value2"
  }
}

variable "object_var" {
  type = object({
    name = string
    age  = number
  })
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate=false")

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify all types are in output
	assert.Contains(t, stdout, `"simple_string"`)
	assert.Contains(t, stdout, `"number_var"`)
	assert.Contains(t, stdout, `"bool_var"`)
	assert.Contains(t, stdout, `"list_var"`)
	assert.Contains(t, stdout, `"map_var"`)
	assert.Contains(t, stdout, `"object_var"`)
}

func TestCLI_SensitiveVariable(t *testing.T) {
	tmpDir := t.TempDir()
	tfFile := filepath.Join(tmpDir, "sensitive.tf")

	tfContent := `
variable "api_key" {
  type      = string
  sensitive = true
}
`
	require.NoError(t, os.WriteFile(tfFile, []byte(tfContent), 0644))

	cmd := setupTestCommand()
	stdout, stderr, err := executeCommand(cmd, "-f", tfFile, "--validate=false")

	require.NoError(t, err, "stderr: %s", stderr)

	// Verify sensitive variable is marked with writeOnly
	assert.Contains(t, stdout, `"api_key"`)
	assert.Contains(t, stdout, `"writeOnly"`)
}

func TestValidateInput(t *testing.T) {
	tests := []struct {
		name      string
		inputDir  string
		inputFile string
		wantErr   bool
		errMsg    string
		setup     func(t *testing.T) string // Returns path to use
	}{
		{
			name:      "no input specified",
			inputDir:  "",
			inputFile: "",
			wantErr:   true,
			errMsg:    "either --dir or --file must be specified",
		},
		{
			name:      "both inputs specified",
			inputDir:  "somedir",
			inputFile: "somefile",
			wantErr:   true,
			errMsg:    "only one of --dir or --file can be specified",
		},
		{
			name:     "valid directory",
			wantErr:  false,
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
		},
		{
			name:    "valid file",
			wantErr: false,
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				tmpFile := filepath.Join(tmpDir, "test.tf")
				require.NoError(t, os.WriteFile(tmpFile, []byte("# test"), 0644))
				return tmpFile
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				path := tt.setup(t)
				if tt.name == "valid directory" {
					inputDir = path
					inputFile = ""
				} else if tt.name == "valid file" {
					inputFile = path
					inputDir = ""
				}
			} else {
				inputDir = tt.inputDir
				inputFile = tt.inputFile
			}
			verbose = false

			// Execute
			err := validateInput()

			// Assert
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}
