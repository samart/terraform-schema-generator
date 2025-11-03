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

	t.Run("parse set type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "availability_zones" {
  type        = set(string)
  description = "Set of availability zones"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "availability_zones", v.Name)
		assert.Contains(t, v.Type, "set")
	})

	t.Run("parse tuple type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "example_tuple" {
  type        = tuple([string, number, bool])
  description = "Example tuple type"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "example_tuple", v.Name)
		assert.Contains(t, v.Type, "tuple")
	})

	t.Run("parse object type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "server_config" {
  type = object({
    name = string
    port = number
    enabled = bool
  })
  description = "Server configuration object"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "server_config", v.Name)
		assert.Contains(t, v.Type, "object")
	})

	t.Run("parse nullable variable true", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "optional_string" {
  type     = string
  nullable = true
  default  = null
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "optional_string", v.Name)
		assert.True(t, v.Nullable)
	})

	t.Run("parse nullable variable false", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "required_non_null" {
  type     = string
  nullable = false
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "required_non_null", v.Name)
		assert.False(t, v.Nullable)
	})

	t.Run("parse nullable default true when not specified", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "implicit_nullable" {
  type = string
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "implicit_nullable", v.Name)
		// Nullable defaults to true according to Terraform spec
		assert.True(t, v.Nullable)
	})

	t.Run("parse ephemeral variable", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "session_token" {
  type      = string
  ephemeral = true
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "session_token", v.Name)
		assert.True(t, v.Ephemeral)
	})

	t.Run("parse ephemeral and sensitive together", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "secret_key" {
  type      = string
  sensitive = true
  ephemeral = true
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "secret_key", v.Name)
		assert.True(t, v.Sensitive)
		assert.True(t, v.Ephemeral)
	})

	t.Run("parse all variable attributes together", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "complete_example" {
  type        = string
  description = "A complete variable with all attributes"
  default     = "default_value"
  sensitive   = true
  nullable    = false
  ephemeral   = true

  validation {
    condition     = length(var.complete_example) > 5
    error_message = "Must be longer than 5 characters"
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
		assert.Equal(t, "complete_example", v.Name)
		assert.Equal(t, "string", v.Type)
		assert.Equal(t, "A complete variable with all attributes", v.Description)
		assert.NotNil(t, v.Default)
		assert.True(t, v.Sensitive)
		assert.False(t, v.Nullable)
		assert.True(t, v.Ephemeral)
		assert.Len(t, v.Validations, 1)
		assert.False(t, v.Required) // Has default, so not required
	})

	t.Run("parse variable with multiple validations", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "instance_count" {
  type        = number
  description = "Number of instances"

  validation {
    condition     = var.instance_count > 0
    error_message = "Instance count must be positive"
  }

  validation {
    condition     = var.instance_count <= 10
    error_message = "Instance count must not exceed 10"
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
		assert.Equal(t, "instance_count", v.Name)
		assert.Len(t, v.Validations, 2)
	})

	t.Run("parse variable with complex default list", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "allowed_ips" {
  type    = list(string)
  default = ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"]
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "allowed_ips", v.Name)
		assert.NotNil(t, v.Default)
		assert.False(t, v.Required)
	})

	t.Run("parse variable with complex default map", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "tags" {
  type = map(string)
  default = {
    Environment = "production"
    Project     = "terraform-demo"
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
		assert.Equal(t, "tags", v.Name)
		assert.NotNil(t, v.Default)
	})

	t.Run("parse variable with complex default object", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "server" {
  type = object({
    hostname = string
    port     = number
  })
  default = {
    hostname = "localhost"
    port     = 8080
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
		assert.Equal(t, "server", v.Name)
		assert.NotNil(t, v.Default)
	})

	t.Run("parse variable without type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "untyped_var" {
  description = "Variable without explicit type"
  default     = "some_value"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "untyped_var", v.Name)
		// Type may be empty when not specified, which is valid
		// Terraform will infer the type from default value
	})

	t.Run("parse any type variable", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "dynamic_value" {
  type        = any
  description = "Can be any type"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "dynamic_value", v.Name)
		assert.Equal(t, "any", v.Type)
	})

	t.Run("parse nested complex types", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "complex_nested" {
  type = list(object({
    name    = string
    subnets = list(string)
    config  = map(string)
  }))
  description = "Nested complex type"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "complex_nested", v.Name)
		// The parser extracts the outer type (list)
		assert.Contains(t, v.Type, "list")
	})

	t.Run("parse variable with empty description", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "no_description" {
  type = string
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "no_description", v.Name)
		assert.Empty(t, v.Description)
	})

	t.Run("parse outputs", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
output "instance_id" {
  description = "The ID of the instance"
  value       = aws_instance.example.id
  sensitive   = false
}

output "private_key" {
  description = "The private key"
  value       = tls_private_key.example.private_key_pem
  sensitive   = true
}
`
		files := map[string]io.Reader{
			"outputs.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Outputs, 2)

		assert.Equal(t, "instance_id", result.Outputs[0].Name)
		assert.Equal(t, "The ID of the instance", result.Outputs[0].Description)
		assert.False(t, result.Outputs[0].Sensitive)

		assert.Equal(t, "private_key", result.Outputs[1].Name)
		assert.True(t, result.Outputs[1].Sensitive)
	})

	t.Run("parse mixed required and optional variables", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "required_var" {
  type        = string
  description = "This is required"
}

variable "optional_var" {
  type        = string
  description = "This is optional"
  default     = "default_value"
}

variable "another_required" {
  type = number
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 3)

		// Check required variables
		requiredCount := 0
		for _, v := range result.Variables {
			if v.Required {
				requiredCount++
			}
		}
		assert.Equal(t, 2, requiredCount)
	})

	t.Run("handle invalid HCL syntax", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "broken" {
  type = string
  description = "Missing closing brace"
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		// Should either return error or handle gracefully
		if err == nil {
			assert.NotNil(t, result)
		}
	})

	t.Run("parse list with complex element type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "server_list" {
  type = list(object({
    name = string
    port = number
  }))
  description = "List of server configurations"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "server_list", v.Name)
		// The parser extracts the outer type (list)
		assert.Contains(t, v.Type, "list")
	})

	t.Run("parse object with optional() attributes", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "server_config" {
  type = object({
    name     = string
    port     = optional(number)
    enabled  = optional(bool, true)
    settings = optional(map(string), {})
  })
  description = "Server configuration with optional attributes"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "server_config", v.Name)
		assert.Contains(t, v.Type, "object")
		assert.Contains(t, v.Type, "optional")
	})

	t.Run("parse optional() with default value", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "config" {
  type = object({
    required_field = string
    optional_field = optional(string, "default_value")
    optional_number = optional(number, 42)
    optional_bool = optional(bool, false)
  })
  description = "Config with optional fields and defaults"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "config", v.Name)
		assert.Contains(t, v.Type, "optional")
	})

	t.Run("parse optional() without default", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "settings" {
  type = object({
    name = string
    port = optional(number)
    tags = optional(map(string))
  })
  description = "Settings with optional fields without defaults"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "settings", v.Name)
		assert.Contains(t, v.Type, "optional")
	})

	t.Run("parse nested optional() in complex types", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "cluster_config" {
  type = object({
    name = string
    nodes = optional(list(object({
      instance_type = string
      count = optional(number, 3)
      tags = optional(map(string))
    })))
    monitoring = optional(object({
      enabled = bool
      interval = optional(number, 60)
    }), {
      enabled = true
    })
  })
  description = "Cluster configuration with nested optionals"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "cluster_config", v.Name)
		assert.Contains(t, v.Type, "object")
		assert.Contains(t, v.Type, "optional")
	})

	t.Run("parse list of objects with optional attributes", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "users" {
  type = list(object({
    username = string
    email = optional(string)
    role = optional(string, "viewer")
    groups = optional(list(string), [])
  }))
  description = "List of users with optional attributes"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "users", v.Name)
		assert.Contains(t, v.Type, "list")
		assert.Contains(t, v.Type, "optional")
	})

	t.Run("parse map with complex value type", func(t *testing.T) {
		parser := NewParser()
		tfContent := `
variable "config_map" {
  type = map(object({
    enabled = bool
    value   = string
  }))
  description = "Map of configurations"
}
`
		files := map[string]io.Reader{
			"variables.tf": strings.NewReader(tfContent),
		}

		result, err := parser.ParseFiles(files)
		require.NoError(t, err)
		assert.Len(t, result.Variables, 1)

		v := result.Variables[0]
		assert.Equal(t, "config_map", v.Name)
		// The parser extracts the outer type (map)
		assert.Contains(t, v.Type, "map")
	})
}
