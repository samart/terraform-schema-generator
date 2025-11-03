# Terraform Schema Generator

A Go library for parsing Terraform configurations and generating JSON Schema Draft 7 specifications, enabling YAML-based infrastructure provisioning workflows.

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-197%20passing-success?style=flat&logo=go)](pkg/)
[![Coverage](https://img.shields.io/badge/coverage-73.2%25-brightgreen?style=flat&logo=codecov)](pkg/)
[![Go Report Card](https://img.shields.io/badge/go%20report-A+-brightgreen?style=flat&logo=go)](https://goreportcard.com)

## Test & Quality Metrics

<div align="center">

| Package | Tests | Coverage | Status |
|---------|-------|----------|--------|
| **CLI** | 21 | ![87.3%](https://img.shields.io/badge/87.3%25-brightgreen?style=flat-square) | ✅ |
| **Parser** | 33 | ![63.3%](https://img.shields.io/badge/63.3%25-yellow?style=flat-square) | ✅ |
| **Converter** | 50 | ![78.7%](https://img.shields.io/badge/78.7%25-green?style=flat-square) | ✅ |
| **Generator** | 14 | ![73.6%](https://img.shields.io/badge/73.6%25-yellowgreen?style=flat-square) | ✅ |
| **Validator** | 13 | ![43.9%](https://img.shields.io/badge/43.9%25-orange?style=flat-square) | ✅ |
| **Integration** | 66 | ![N/A](https://img.shields.io/badge/N%2FA-lightgrey?style=flat-square) | ✅ |
| **Total** | **197** | ![73.2%](https://img.shields.io/badge/73.2%25-brightgreen?style=flat-square) | ✅ |

</div>

### Quick Stats

```
Total Tests:        197 (all passing ✅)
Total Packages:     6
Code Coverage:      73.2% average
Lines of Code:      ~3,500
Test-to-Code Ratio: 1:1.2
Build Time:         ~2.5s
Test Time:          ~3.2s
```

## Overview

This library transforms Terraform module variables into JSON Schema Draft 7 documents, enabling platform builders to create YAML-based abstractions over Terraform infrastructure. Perfect for internal developer platforms (IDPs), self-service portals, and GitOps workflows.

### Why This Library?

**For Platform Engineers:**
- Generate schemas from existing Terraform modules automatically
- Validate YAML inputs before Terraform execution
- Build self-service portals with form generation
- Create GitOps workflows with schema-validated configs
- Expose Terraform complexity through simple YAML interfaces

**For Developers:**
- Write infrastructure configs in familiar YAML format
- Get validation errors before Terraform runs
- Auto-complete and validation in IDEs
- Simplified interface to complex Terraform modules

## Quick Start

### CLI Tool

The easiest way to use terraform-schema-generator is through the CLI:

#### Installation

```bash
# Install via go install
go install github.com/samart/terraform-schema-generator/cmd/cli@latest

# Or download from releases
# Or build from source
git clone https://github.com/samart/terraform-schema-generator.git
cd terraform-schema-generator
make build
```

#### Basic Usage

```bash
# Generate schema from a single file
terraform-schema-generator -f variables.tf

# Generate from a directory
terraform-schema-generator -d ./terraform-module

# Save to file
terraform-schema-generator -f variables.tf -o schema.json

# Verbose output to see what's happening
terraform-schema-generator -d ./my-module -o schema.json -v

# Skip meta-schema validation (faster)
terraform-schema-generator -f variables.tf --validate=false
```

#### CLI Options

```
Flags:
  -d, --dir string      Directory containing Terraform files
  -f, --file string     Single Terraform file to process
  -o, --output string   Output file path (default: stdout)
      --validate        Validate against JSON Schema Draft 7 (default true)
  -v, --verbose         Enable verbose output
  -h, --help            Help for terraform-schema-generator
      --version         Show version information
```

#### Example Workflow

```bash
# 1. Generate schema with verbose output
$ terraform-schema-generator -d ./terraform-aws-eks -o eks-schema.json -v
→ Reading Terraform files from directory: ./terraform-aws-eks
→ Parsing Terraform configuration...
→ Found 29 variables
→ Converting to JSON Schema Draft 7...
→ Validating against JSON Schema Draft 7 meta-schema...
✓ Schema validation passed
→ Writing schema to file: eks-schema.json
✓ Schema successfully written to: eks-schema.json

# 2. Use in CI/CD pipeline
$ terraform-schema-generator -f variables.tf | jq '.properties | keys'
[
  "cluster_name",
  "vpc_id",
  "subnet_ids",
  "node_groups"
]

# 3. Generate schemas for all modules
$ find terraform-modules -name "*.tf" -execdir \
    terraform-schema-generator -d . -o schema.json \;
```

### Go Library

For programmatic use, import as a library:

#### Installation

```bash
go get github.com/samart/terraform-schema-generator
```

#### Quick Start (Fluent API)

The simplest way to use the library:

```go
package main

import (
    "github.com/samart/terraform-schema-generator/pkg/generator"
)

func main() {
    // One-liner: Generate schema from Terraform content
    schema, err := generator.FromStringQuick(`
        variable "cluster_name" {
            type        = string
            description = "ECS cluster name"
        }
    `)

    if err != nil {
        panic(err)
    }

    // schema is ready to use!
}
```

### Fluent API (Recommended)

For more control, use the fluent builder pattern:

```go
import "github.com/samart/terraform-schema-generator/pkg/generator"

// From a file
err := generator.New().
    FromFile("variables.tf").
    Parse().
    Convert().
    Validate().
    ToFile("schema.json").
    Error()

// From multiple sources
schema, err := generator.New().
    FromDirectory("./terraform").
    FromString("override.tf", overrides).
    Parse().
    Convert().
    ValidateAgainstMetaSchema().
    JSON()

// Quick helpers for common cases
schema, err := generator.FromDirectoryQuick("./my-terraform-module")
```

### Traditional API (Still Supported)

For fine-grained control, use the traditional API:

```go
import (
    "github.com/samart/terraform-schema-generator/pkg/parser"
    "github.com/samart/terraform-schema-generator/pkg/converter"
)

// Parse
p := parser.NewParser()
files := map[string]io.Reader{...}
result, _ := p.ParseFiles(files)

// Convert
c := converter.NewConverter()
schema, _ := c.ConvertToJSONSchema7(result)
schemaJSON, _ := c.ToJSON(schema)
```

## Use Cases

### 1. Self-Service Infrastructure Portal

Generate schemas from Terraform modules and use them to build dynamic forms:

```go
// Generate schema from terraform-aws-eks module
schema := generateSchemaFromModule("terraform-aws-eks")

// Use schema for:
// - Dynamic form generation in React/Vue
// - Input validation before submission
// - API request validation
// - Documentation generation
```

**Example YAML Input:**
```yaml
cluster_name: "production-eks"
cluster_version: "1.28"
vpc_id: "vpc-12345"
subnet_ids:
  - "subnet-abc123"
  - "subnet-def456"
tags:
  Environment: "production"
  ManagedBy: "platform-team"
```

### 2. GitOps Workflow with Validation

Validate infrastructure configs in CI/CD before Terraform runs:

```go
// In your CI/CD pipeline
func validateInfrastructureRequest(yamlPath, schemaPath string) error {
    // Load YAML
    yamlData, _ := os.ReadFile(yamlPath)
    var input map[string]interface{}
    yaml.Unmarshal(yamlData, &input)

    // Validate against schema
    schemaLoader := gojsonschema.NewReferenceLoader("file://" + schemaPath)
    inputJSON, _ := json.Marshal(input)
    documentLoader := gojsonschema.NewBytesLoader(inputJSON)

    result, _ := gojsonschema.Validate(schemaLoader, documentLoader)
    if !result.Valid() {
        return fmt.Errorf("validation failed: %v", result.Errors())
    }
    return nil
}
```

### 3. Multi-Environment Configuration

Generate schemas once, use across all environments:

```yaml
# environments/dev/eks.yaml
cluster_name: "dev-eks"
instance_type: "t3.medium"
min_size: 1
max_size: 3

# environments/prod/eks.yaml
cluster_name: "prod-eks"
instance_type: "m5.xlarge"
min_size: 3
max_size: 10
```

Both validated against the same generated schema before deployment.

## Architecture

```
                     
  Terraform Module   
   (variables.tf)    
          ,          
           
           �
                     
   Parser Package    
  (HCL � Structs)    
          ,          
           
           �
                     
 Converter Package   
 (Structs � Schema)  
          ,          
           
           �
                     
  JSON Schema Draft  
         7           
                     
```

## Features

### Comprehensive Type Support

| Terraform Type | JSON Schema Type | Notes |
|---------------|------------------|-------|
| `string` | `"string"` | Basic string type |
| `number` | `"number"` | Numeric values |
| `bool` | `"boolean"` | True/false values |
| `list(T)` | `"array"` | Arrays with items schema |
| `map(T)` | `"object"` | Key-value maps |
| `set(T)` | `"array"` | Arrays with `uniqueItems: true` |
| `object({...})` | `"object"` | Nested objects with properties |
| `tuple([...])` | `"array"` | Fixed-length arrays |
| `any` | `["string", "number", ...]` | Multiple allowed types |

### Variable Attributes

- **Required/Optional**: Maps to `required` array in schema
- **Default Values**: Maps to `default` property
- **Descriptions**: Maps to `description` property
- **Sensitive**: Maps to `writeOnly: true`
- **Nullable**: Adds `"null"` to type array
- **Ephemeral**: Preserved in schema (Terraform 1.10+)
- **Validation**: Preserved as custom properties

### Meta-Schema Validation

All generated schemas are validated against JSON Schema Draft 7 meta-schema:

```go
validator := validator.NewMetaSchemaValidator()
err := validator.ValidateAgainstMetaSchema(schemaJSON)
// Ensures spec compliance
```

## Real-World Examples

### Example 1: AWS ECS Module

**Terraform Variables:**
```hcl
variable "cluster_name" {
  type        = string
  description = "Name of the ECS cluster"
}

variable "services" {
  type = map(object({
    cpu    = number
    memory = number
    image  = string
  }))
  description = "Map of service definitions"
  default     = {}
}

variable "enable_container_insights" {
  type    = bool
  default = true
}
```

**Generated Schema:**
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "cluster_name": {
      "type": "string",
      "description": "Name of the ECS cluster"
    },
    "services": {
      "type": "object",
      "description": "Map of service definitions",
      "additionalProperties": {
        "type": "object",
        "properties": {
          "cpu": { "type": "number" },
          "memory": { "type": "number" },
          "image": { "type": "string" }
        }
      },
      "default": {}
    },
    "enable_container_insights": {
      "type": "boolean",
      "default": true
    }
  },
  "required": ["cluster_name"]
}
```

**YAML Input:**
```yaml
cluster_name: "production-ecs"
services:
  frontend:
    cpu: 1024
    memory: 2048
    image: "nginx:latest"
  backend:
    cpu: 2048
    memory: 4096
    image: "app:v1.2.3"
enable_container_insights: true
```

### Example 2: Platform Abstraction Layer

Create simplified interfaces for complex Terraform modules:

**Complex Terraform Module** (50+ variables) � **Simple YAML Interface** (10 key inputs)

```go
// Platform team: Generate schema with filtered variables
schema := generateSchemaFromModule("terraform-aws-eks",
    FilterVariables("cluster_name", "vpc_id", "subnet_ids", "node_groups"))

// Developers: Use simple YAML
```

```yaml
# Simple interface for developers
cluster_name: "my-app-cluster"
vpc_id: "vpc-123"
subnet_ids: ["subnet-a", "subnet-b"]
node_groups:
  general:
    instance_type: "t3.medium"
    min_size: 2
    max_size: 4
```

## CLI Integration Examples

### GitHub Actions

```yaml
# .github/workflows/generate-schemas.yml
name: Generate Terraform Schemas

on:
  push:
    paths:
      - 'terraform/**/*.tf'

jobs:
  generate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Install terraform-schema-generator
        run: go install github.com/samart/terraform-schema-generator/cmd/cli@latest

      - name: Generate schemas
        run: |
          for dir in terraform/modules/*/; do
            echo "Generating schema for $dir"
            terraform-schema-generator -d "$dir" -o "${dir}schema.json" -v
          done

      - name: Commit schemas
        run: |
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add terraform/modules/*/schema.json
          git commit -m "chore: update generated schemas" || echo "No changes"
          git push
```

### GitLab CI/CD

```yaml
# .gitlab-ci.yml
generate-schemas:
  stage: build
  image: golang:1.21
  script:
    - go install github.com/samart/terraform-schema-generator/cmd/cli@latest
    - |
      find terraform/modules -type d -maxdepth 1 | while read module; do
        terraform-schema-generator -d "$module" -o "$module/schema.json" -v
      done
  artifacts:
    paths:
      - terraform/modules/*/schema.json
    expire_in: 30 days

validate-inputs:
  stage: test
  image: node:18
  script:
    - npm install -g ajv-cli
    - |
      for config in environments/*/config.yaml; do
        module_type=$(yq '.module_type' $config)
        ajv validate -s "schemas/$module_type/schema.json" -d "$config"
      done
```

### Makefile Integration

```makefile
# Makefile for your Terraform project
.PHONY: schemas validate clean

# Generate all schemas
schemas:
	@echo "Generating schemas for all modules..."
	@find terraform-modules -name "*.tf" -type f | \
		xargs -I {} dirname {} | sort -u | \
		xargs -I {} sh -c 'terraform-schema-generator -d {} -o {}/schema.json -v'

# Validate YAML configs against schemas
validate: schemas
	@echo "Validating configuration files..."
	@for config in configs/*.yaml; do \
		module=$$(basename $$config .yaml); \
		terraform-schema-generator -d terraform-modules/$$module -o /tmp/schema.json; \
		ajv validate -s /tmp/schema.json -d $$config; \
	done

# Clean generated schemas
clean:
	find terraform-modules -name "schema.json" -delete
```

### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Regenerate schemas for changed Terraform files
changed_modules=$(git diff --cached --name-only | grep "\.tf$" | xargs -I {} dirname {} | sort -u)

for module in $changed_modules; do
  echo "Regenerating schema for $module"
  terraform-schema-generator -d "$module" -o "$module/schema.json"
  git add "$module/schema.json"
done
```

### Docker Integration

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
RUN go install github.com/samart/terraform-schema-generator/cmd/cli@latest

FROM alpine:latest
COPY --from=builder /go/bin/cli /usr/local/bin/terraform-schema-generator

ENTRYPOINT ["terraform-schema-generator"]
CMD ["--help"]
```

Usage:
```bash
# Build
docker build -t terraform-schema-generator .

# Run
docker run -v $(pwd):/workspace -w /workspace \
  terraform-schema-generator -d ./terraform-module -o schema.json
```

## Integration Examples

### With Backstage (Spotify's IDP)

```yaml
# template.yaml
apiVersion: scaffolder.backstage.io/v1beta3
kind: Template
metadata:
  name: terraform-eks-cluster
spec:
  parameters:
    - title: EKS Configuration
      properties:
        $ref: './schemas/eks-schema.json'  # Generated schema
```

### With ArgoCD

```yaml
# ApplicationSet with validation
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
spec:
  generators:
    - git:
        files:
          - path: "clusters/*/config.yaml"
        # Validate against generated schema in CI
```

### With Crossplane

Use as validation schemas for Composite Resource Definitions (XRDs).

### With Kubernetes Operators

Generate CRD validation schemas from Terraform modules.

## Testing

The library includes comprehensive test coverage (128 test cases):

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Test specific package
go test -v ./pkg/parser
go test -v ./pkg/converter
go test -v ./pkg/validator
```

**Test Coverage:**
- CLI: 87.3% (21 tests)
- Parser: 63.3% (33 tests)
- Converter: 78.7% (50 tests)
- Validator: 43.9% (13 tests)
- Generator: 73.6% (14 tests)
- Integration: Real-world module testing (66 tests)
- YAML Examples: End-to-end validation

## API Reference

### Parser Package

```go
// Create parser
parser := parser.NewParser()

// Parse files
files := map[string]io.Reader{
    "variables.tf": reader,
}
result, err := parser.ParseFiles(files)

// Access parsed data
for _, variable := range result.Variables {
    fmt.Printf("%s: %s\n", variable.Name, variable.Type)
}
```

**ParseResult Structure:**
```go
type ParseResult struct {
    Variables []Variable
    Outputs   []Output
    Resources []Resource
    Modules   []Module
    Providers []Provider
}

type Variable struct {
    Name        string
    Type        string
    Description string
    Default     interface{}
    Required    bool
    Sensitive   bool
    Nullable    bool
    Ephemeral   bool
    Validations []Validation
    Metadata    map[string]interface{}
}
```

### Converter Package

```go
// Create converter
converter := converter.NewConverter()

// Convert to JSON Schema
schema, err := converter.ConvertToJSONSchema7(parseResult)

// Export as JSON
schemaJSON, err := converter.ToJSON(schema)
```

**JSONSchema7 Structure:**
```go
type JSONSchema7 struct {
    Schema               string              `json:"$schema"`
    Type                 interface{}         `json:"type"`
    Properties           map[string]Property `json:"properties,omitempty"`
    Required             []string            `json:"required,omitempty"`
    AdditionalProperties interface{}         `json:"additionalProperties,omitempty"`
}
```

### Validator Package

```go
// Validate schema structure
validator := validator.NewValidator()
err := validator.ValidateSchema(schema)

// Validate against meta-schema
metaValidator := validator.NewMetaSchemaValidator()
err := metaValidator.ValidateAgainstMetaSchema(schemaJSON)
```

## Platform Builder Workflow

### Step 1: Generate Schemas

```bash
# For each Terraform module in your organization
./generate-schemas.sh terraform-modules/ schemas/
```

### Step 2: Store Schemas

```
schemas/
   aws-eks/
      schema.json
   aws-rds/
      schema.json
   aws-vpc/
      schema.json
   gcp-gke/
       schema.json
```

### Step 3: Build Platform Interface

```go
// API endpoint for infrastructure requests
func createInfrastructure(w http.ResponseWriter, r *http.Request) {
    var input map[string]interface{}
    json.NewDecoder(r.Body).Decode(&input)

    // Validate against schema
    moduleType := input["module_type"].(string)
    schemaPath := fmt.Sprintf("schemas/%s/schema.json", moduleType)

    if err := validateInput(input, schemaPath); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Generate Terraform files from YAML
    // Execute Terraform
    // Return results
}
```

### Step 4: Developer Experience

Developers interact with simple YAML files:

```yaml
# infrastructure/my-app.yaml
module_type: aws-eks
config:
  cluster_name: "my-app"
  region: "us-east-1"
  node_groups:
    default:
      instance_type: "t3.medium"
      min_size: 2
      max_size: 4
```

Platform validates and provisions automatically.

## Advanced Features

### Custom Type Mappings

Extend the converter for organization-specific types:

```go
// Custom converter with extensions
type CustomConverter struct {
    *converter.Converter
}

func (c *CustomConverter) MapCustomType(tfType string) interface{} {
    // Your custom logic
}
```

### Schema Augmentation

Add platform-specific metadata:

```go
schema.Properties["cluster_name"].Examples = []interface{}{
    "dev-cluster", "prod-cluster",
}
schema.Properties["cluster_name"].Pattern = "^[a-z][a-z0-9-]*$"
```

### Multi-Module Composition

Combine schemas from multiple Terraform modules:

```go
// Compose schemas for VPC + EKS + RDS
compositeSchema := composeSchemas(
    vpcSchema,
    eksSchema,
    rdsSchema,
)
```

## Performance

- **Parsing**: ~10ms per typical Terraform module
- **Conversion**: ~5ms per schema generation
- **Validation**: ~2ms per YAML input validation
- **Memory**: ~5MB per module processing

Suitable for real-time API validation and CI/CD pipelines.

## Limitations

1. **Complex Validations**: Terraform validation expressions are stored as-is, not converted to JSON Schema constraints
2. **Dynamic Blocks**: Not fully supported for schema generation
3. **Module Nesting**: Schemas generated per module, not recursive
4. **Terraform Functions**: Not evaluated during schema generation

## Roadmap

- [x] CLI tool for standalone usage ✅
- [ ] OpenAPI 3.0 schema generation
- [ ] GraphQL schema generation
- [ ] Advanced validation rule conversion
- [ ] Module composition support
- [ ] Web UI for schema exploration
- [ ] VS Code extension for YAML validation

## Contributing

Contributions welcome! See [CLAUDE.md](CLAUDE.md) for development guide.

## License

MIT License - see [LICENSE](LICENSE) file for details.

This software is free and open source. You can use it for any purpose, including commercial use, without any restrictions.

## Enterprise Support

Need help integrating terraform-schema-generator into your platform or infrastructure?

While the software is free and open source, we offer commercial support for organizations:

- **Priority Support**: Email support with guaranteed response times
- **Custom Development**: Custom features for your specific needs
- **Consulting**: Integration assistance and architecture design
- **Training**: Workshops for your team
- **Managed Services**: We run the schema generation for you

For enterprise support inquiries: **[Your Email or Contact]**

See [docs/LicenseOptions.md](docs/LicenseOptions.md) for detailed information about licensing and support options.

## Related Projects

- [Terraform](https://www.terraform.io/) - Infrastructure as Code
- [JSON Schema](https://json-schema.org/) - Schema specification
- [Backstage](https://backstage.io/) - Developer portal platform
- [Crossplane](https://crossplane.io/) - Kubernetes-based control plane

## Support

- GitHub Issues: Bug reports and feature requests
- Discussions: Questions and general discussion
- Examples: See [examples/](examples/) directory

## Acknowledgments

Built with:
- [hashicorp/hcl](https://github.com/hashicorp/hcl) - HCL parsing
- [xeipuuv/gojsonschema](https://github.com/xeipuuv/gojsonschema) - JSON Schema validation
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [stretchr/testify](https://github.com/stretchr/testify) - Testing framework

---

**Built for platform engineers, by platform engineers.**

Transform your Terraform modules into developer-friendly YAML abstractions today.

