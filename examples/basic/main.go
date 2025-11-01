package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/samart/terraform-schema-generator/pkg/converter"
	"github.com/samart/terraform-schema-generator/pkg/parser"
	"github.com/samart/terraform-schema-generator/pkg/validator"
)

// Example Terraform module configuration
const terraformExample = `
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

variable "instance_count" {
  type        = number
  description = "Number of instances to create"
  default     = 1
}

variable "enable_monitoring" {
  type        = bool
  description = "Enable CloudWatch monitoring"
  default     = false
}

variable "vpc_id" {
  type        = string
  description = "VPC ID for deployment"
}

variable "subnet_ids" {
  type        = list(string)
  description = "List of subnet IDs"
}

variable "tags" {
  type        = map(string)
  description = "Resource tags"
  default     = {}
}

variable "db_password" {
  type        = string
  description = "Database password"
  sensitive   = true
}

output "instance_ids" {
  description = "List of EC2 instance IDs"
  value       = aws_instance.app[*].id
}

output "load_balancer_dns" {
  description = "DNS name of the load balancer"
  value       = aws_lb.main.dns_name
  sensitive   = false
}
`

func main() {
	fmt.Println("🚀 Terraform Schema Generator - Example Usage\n")

	// Step 1: Create parser
	fmt.Println("📋 Step 1: Creating parser...")
	p := parser.NewParser()

	// Step 2: Parse Terraform files
	fmt.Println("📖 Step 2: Parsing Terraform configuration...")
	files := map[string]io.Reader{
		"main.tf": strings.NewReader(terraformExample),
	}

	result, err := p.ParseFiles(files)
	if err != nil {
		log.Fatalf("❌ Failed to parse files: %v", err)
	}

	// Display parsing results
	fmt.Printf("✅ Successfully parsed Terraform files\n")
	fmt.Printf("   - Variables found: %d\n", len(result.Variables))
	fmt.Printf("   - Outputs found: %d\n", len(result.Outputs))
	fmt.Printf("   - Resources found: %d\n", len(result.Resources))
	fmt.Printf("   - Modules found: %d\n\n", len(result.Modules))

	// Display variables
	fmt.Println("📦 Variables:")
	for _, v := range result.Variables {
		requiredStatus := "optional"
		if v.Required {
			requiredStatus = "required"
		}
		sensitiveStatus := ""
		if v.Sensitive {
			sensitiveStatus = " [SENSITIVE]"
		}
		fmt.Printf("   - %s (%s, %s)%s\n", v.Name, v.Type, requiredStatus, sensitiveStatus)
		if v.Description != "" {
			fmt.Printf("     %s\n", v.Description)
		}
	}
	fmt.Println()

	// Display outputs
	if len(result.Outputs) > 0 {
		fmt.Println("📤 Outputs:")
		for _, o := range result.Outputs {
			fmt.Printf("   - %s: %s\n", o.Name, o.Description)
		}
		fmt.Println()
	}

	// Step 3: Convert to JSON Schema
	fmt.Println("🔄 Step 3: Converting to JSON Schema Draft 7...")
	c := converter.NewConverter()

	schema, err := c.ConvertToJSONSchema7(result)
	if err != nil {
		log.Fatalf("❌ Failed to convert to JSON Schema: %v", err)
	}

	fmt.Printf("✅ Conversion successful\n")
	fmt.Printf("   - Schema type: %s\n", schema.Type)
	fmt.Printf("   - Properties: %d\n", len(schema.Properties))
	fmt.Printf("   - Required fields: %d\n\n", len(schema.Required))

	// Step 4: Validate the schema
	fmt.Println("✓ Step 4: Validating generated schema...")
	v := validator.NewValidator()

	if err := v.ValidateSchema(schema); err != nil {
		log.Fatalf("❌ Schema validation failed: %v", err)
	}

	fmt.Println("✅ Schema is valid!\n")

	// Step 5: Output the schema
	fmt.Println("📄 Step 5: Generating JSON Schema output...\n")

	schemaJSON, err := c.ToJSON(schema)
	if err != nil {
		log.Fatalf("❌ Failed to convert schema to JSON: %v", err)
	}

	fmt.Println("Generated JSON Schema:")
	fmt.Println("─────────────────────────────────────────────")
	fmt.Println(string(schemaJSON))
	fmt.Println("─────────────────────────────────────────────")

	// Optional: Save to file
	fmt.Println("\n💾 Saving schema to file...")
	outputFile := "terraform-variables-schema.json"
	if err := os.WriteFile(outputFile, schemaJSON, 0644); err != nil {
		log.Fatalf("❌ Failed to write schema to file: %v", err)
	}

	fmt.Printf("✅ Schema saved to: %s\n", outputFile)

	// Display usage information
	fmt.Println("\n🎯 Usage Information:")
	fmt.Println("───────────────────────────────────────────────")
	fmt.Println("You can now use this schema to:")
	fmt.Println("  • Validate Terraform variable inputs")
	fmt.Println("  • Generate API documentation")
	fmt.Println("  • Create web forms for Terraform modules")
	fmt.Println("  • Integrate with configuration management tools")
	fmt.Println("  • Provide IDE autocomplete support")
	fmt.Println()

	// Display example validation usage
	fmt.Println("📚 Example: Using the schema for validation")
	fmt.Println("───────────────────────────────────────────────")
	fmt.Println(`
// In your application:
import "encoding/json"
import "github.com/xeipuuv/gojsonschema"

// Load the schema
schemaLoader := gojsonschema.NewReferenceLoader("file://./terraform-variables-schema.json")

// Your Terraform variables
documentLoader := gojsonschema.NewStringLoader(` + "`" + `{
  "environment": "prod",
  "vpc_id": "vpc-123456",
  "subnet_ids": ["subnet-1", "subnet-2"],
  "db_password": "secretpass123"
}` + "`" + `)

// Validate
result, err := gojsonschema.Validate(schemaLoader, documentLoader)
if err != nil {
    panic(err)
}

if result.Valid() {
    fmt.Println("✅ Variables are valid!")
} else {
    for _, desc := range result.Errors() {
        fmt.Printf("❌ %s\n", desc)
    }
}
`)

	fmt.Println("\n✨ Example completed successfully!")
}
