package main

import (
	"fmt"
	"log"

	"github.com/samart/terraform-schema-generator/pkg/generator"
)

const terraformConfig = `
variable "environment" {
  type        = string
  description = "Environment name (dev, staging, prod)"

  validation {
    condition     = contains(["dev", "staging", "prod"], var.environment)
    error_message = "Environment must be one of: dev, staging, prod"
  }
}

variable "instance_type" {
  type        = string
  description = "EC2 instance type"
  default     = "t2.micro"
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for deployment"
}

variable "tags" {
  type        = map(string)
  description = "Resource tags"
  default     = {}
}

variable "server_config" {
  type = object({
    name     = string
    port     = optional(number, 8080)
    enabled  = optional(bool, true)
    settings = optional(map(string), {})
  })
  description = "Server configuration with optional attributes"
}
`

func main() {
	fmt.Println("🚀 Terraform Schema Generator - Fluent API Example")
	fmt.Println()

	// Example 1: Simple fluent chain
	fmt.Println("📝 Example 1: Simple fluent chain")
	fmt.Println("─────────────────────────────────────────")

	schemaJSON, err := generator.New().
		FromString("variables.tf", terraformConfig).
		Parse().
		Convert().
		Validate().
		JSON()

	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Generated schema (%d bytes)\n\n", len(schemaJSON))

	// Example 2: With meta-schema validation and file output
	fmt.Println("📝 Example 2: With meta-schema validation and file output")
	fmt.Println("─────────────────────────────────────────")

	err = generator.New().
		FromString("main.tf", terraformConfig).
		Parse().
		Convert().
		Validate().
		ValidateAgainstMetaSchema().
		ToFile("schema.json").
		Error()

	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Println("✅ Schema validated and saved to schema.json")
	fmt.Println()

	// Example 3: Quick helper for one-liner
	fmt.Println("📝 Example 3: One-liner with quick helper")
	fmt.Println("─────────────────────────────────────────")

	schema, err := generator.FromStringQuick(terraformConfig)
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Schema generated with one line (%d bytes)\n\n", len(schema))

	// Example 4: Access intermediate results
	fmt.Println("📝 Example 4: Access parse results before conversion")
	fmt.Println("─────────────────────────────────────────")

	gen := generator.New().
		FromString("terraform.tf", terraformConfig).
		Parse()

	parseResult, err := gen.ParseResult()
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Found %d variables:\n", len(parseResult.Variables))
	for _, v := range parseResult.Variables {
		status := "optional"
		if v.Required {
			status = "required"
		}
		fmt.Printf("   • %s (%s) - %s\n", v.Name, v.Type, status)
	}

	// Continue the chain
	finalSchema, err := gen.Convert().Validate().Schema()
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("\n✅ Converted to schema with %d properties\n\n", len(finalSchema.Properties))

	// Example 5: From directory
	fmt.Println("📝 Example 5: From directory (if testdata exists)")
	fmt.Println("─────────────────────────────────────────")

	// This would work if you have a directory with .tf files
	// gen = generator.New().FromDirectory("./my-terraform-module")

	// Example with error handling
	gen = generator.New().FromDirectory("./nonexistent")
	if err := gen.Error(); err != nil {
		fmt.Printf("⚠️  Expected error for nonexistent directory: %v\n\n", err)
	}

	// Example 6: Multiple files
	fmt.Println("📝 Example 6: Multiple files")
	fmt.Println("─────────────────────────────────────────")

	variables := `
variable "app_name" {
  type = string
}
`

	outputs := `
output "app_url" {
  value = "https://example.com"
}
`

	schema, err = generator.New().
		FromString("variables.tf", variables).
		FromString("outputs.tf", outputs).
		Parse().
		Convert().
		JSON()

	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	fmt.Printf("✅ Generated schema from multiple files (%d bytes)\n\n", len(schema))

	// Summary
	fmt.Println("✨ Fluent API Examples Completed!")
	fmt.Println("═════════════════════════════════════════")
	fmt.Println("\n💡 Key Benefits:")
	fmt.Println("  • Chainable methods for clean, readable code")
	fmt.Println("  • Error handling at the end of the chain")
	fmt.Println("  • Quick helpers for common use cases")
	fmt.Println("  • Access intermediate results when needed")
	fmt.Println("  • Support for files, strings, readers, and directories")
	fmt.Println()
}
