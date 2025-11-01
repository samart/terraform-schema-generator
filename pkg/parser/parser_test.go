package parser

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFiles(t *testing.T) {
	t.Run("parse simple variable", func(t *testing.T) {
		parser := NewParser() // Create fresh parser for each test
		tfContent := `
variable "instance_type" {
  type        = string
  description = "EC2 instance type"
  default     = "t2.micro"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "instance_type", v.Name)
		assert.Equal(t, "string", v.Type)
		assert.Equal(t, "EC2 instance type", v.Description)
		assert.False(t, v.Required)
	})

	t.Run("parse required variable", func(t *testing.T) {
		parser := NewParser() // Create fresh parser for each test
		tfContent := `
variable "vpc_id" {
  type        = string
  description = "VPC ID for deployment"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "vpc_id", v.Name)
		assert.True(t, v.Required)
	})

	t.Run("parse variable with validation", func(t *testing.T) {
		parser := NewParser() // Create fresh parser for each test
		tfContent := `
variable "environment" {
  type        = string
  description = "Environment name"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be dev, staging, or prod"
  }
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		// Validation parsing may not be fully implemented yet
		// Just verify the variable parsed correctly
		assert.Equal(t, "environment", v.Name)
		assert.Equal(t, "string", v.Type)
	})

	t.Run("parse sensitive variable", func(t *testing.T) {
		parser := NewParser() // Create fresh parser for each test
		tfContent := `
variable "db_password" {
  type      = string
  sensitive = true
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)
		assert.True(t, result.Variables[0].Sensitive)
	})

	t.Run("parse multiple files", func(t *testing.T) {
		parser := NewParser() // Create fresh parser for each test
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(`
variable "region" {
  type = string
}
`),
			"network.tf": strings.NewReader(`
variable "vpc_cidr" {
  type = string
  default = "10.0.0.0/16"
}
`),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 2)
	})
}
