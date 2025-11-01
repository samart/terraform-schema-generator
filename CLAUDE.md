# Terraform Schema Generator - Development Guide

## Overview

This is a reusable Go library for parsing Terraform configurations and generating JSON Schema Draft 7 specifications. Extracted from the iac-platform project to be a standalone, reusable component.

## Development Environment

### Prerequisites

- Go 1.21 or higher
- Git
- Understanding of Terraform HCL syntax
- Familiarity with JSON Schema Draft 7

### Setup

```bash
# Clone the repository
git clone https://github.com/samart/terraform-schema-generator.git
cd terraform-schema-generator

# Download dependencies
go mod download

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Architecture

### Package Structure

```
pkg/
├── parser/          # Terraform HCL parsing and extraction
├── converter/       # JSON Schema 7 conversion
└── validator/       # Schema validation
```

### Design Principles

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Testability**: All packages are thoroughly tested with unit tests
3. **Reusability**: Designed to be imported and used in other projects
4. **Type Safety**: Strong typing throughout the codebase
5. **Error Handling**: Comprehensive error handling and reporting

## Core Components

### 1. Parser Package (`pkg/parser`)

**Purpose**: Parse Terraform HCL files and extract structured data

**Key Types**:
- `Parser`: Main parsing engine using HashiCorp's HCL library
- `ParseResult`: Container for all extracted Terraform components
- `Variable`, `Output`, `Resource`, `Module`, `Provider`: Structured representations

**Responsibilities**:
- Parse multiple Terraform files
- Extract variables with all attributes (type, description, default, validation, etc.)
- Extract outputs, resources, modules, and provider requirements
- Handle parsing errors gracefully

### 2. Converter Package (`pkg/converter`)

**Purpose**: Convert parsed Terraform structures to JSON Schema Draft 7

**Key Types**:
- `Converter`: Conversion engine
- `JSONSchema7`: JSON Schema document structure
- `Property`: JSON Schema property definition

**Type Mapping Logic**:
```go
Terraform Type → JSON Schema Type
string         → "string"
number         → "number"
bool           → "boolean"
list(T)        → "array"
map(T)         → "object"
object({...})  → "object"
any            → ["string", "number", "boolean", "object", "array", "null"]
```

**Features**:
- Maps Terraform types to JSON Schema types
- Preserves validation rules
- Handles sensitive variables (writeOnly)
- Maintains required/optional distinction

### 3. Validator Package (`pkg/validator`)

**Purpose**: Validate generated JSON schemas

**Validations**:
- Schema version compliance (Draft 7)
- Required field presence
- Type validity
- Property constraints (min/max, length, etc.)
- Schema marshaling capability

## Testing Strategy

### Test Coverage Goals

- **Target**: >80% code coverage
- **Critical Paths**: 100% coverage on parsing and conversion logic
- **Edge Cases**: All error conditions tested

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    t.Run("happy path", func(t *testing.T) {
        // Test normal operation
    })

    t.Run("edge case", func(t *testing.T) {
        // Test boundary conditions
    })

    t.Run("error handling", func(t *testing.T) {
        // Test error scenarios
    })
}
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./pkg/parser -v

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# With race detection
go test -race ./...
```

## Development Workflow

### 1. Making Changes

```bash
# Create feature branch
git checkout -b feature/your-feature

# Make changes
# Write tests first (TDD approach recommended)

# Run tests
go test ./...

# Commit with clear message
git commit -m "feat: add support for complex object types"
```

### 2. Code Style

Follow standard Go conventions:
- `gofmt` for formatting
- Clear, descriptive variable names
- Comments for exported functions
- Error messages should be lowercase and not end with punctuation

### 3. Adding New Features

Example: Adding support for a new Terraform type

1. **Update Parser**: Add extraction logic in `parser/parser.go`
2. **Update Converter**: Add type mapping in `converter/converter.go`
3. **Add Tests**: Create test cases for the new feature
4. **Update Documentation**: Document the new capability

## Common Tasks

### Adding a New Terraform Component Type

```go
// 1. Define the structure in parser/parser.go
type NewComponent struct {
    Name string `json:"name"`
    // ... other fields
}

// 2. Add to ParseResult
type ParseResult struct {
    Variables []Variable
    NewComponents []NewComponent // Add here
}

// 3. Add extraction method
func (p *Parser) extractNewComponents(file *hcl.File) []NewComponent {
    // Extraction logic
}

// 4. Call in ParseFiles
func (p *Parser) ParseFiles(files map[string]io.Reader) (*ParseResult, error) {
    // ... existing code
    newComps := p.extractNewComponents(file)
    result.NewComponents = append(result.NewComponents, newComps...)
}
```

### Extending Type Mapping

```go
// In converter/converter.go
func (c *Converter) mapTerraformTypeToJSONSchema(tfType string) interface{} {
    // Add new type mapping
    if strings.HasPrefix(tfType, "your_new_type") {
        return "your_json_schema_type"
    }
    // ... existing mappings
}
```

## Dependencies

### Direct Dependencies

```go
require (
    github.com/hashicorp/hcl/v2 v2.19.1  // HCL parsing
    github.com/stretchr/testify v1.8.4   // Testing utilities
)
```

### Updating Dependencies

```bash
# Update all dependencies
go get -u ./...

# Update specific dependency
go get -u github.com/hashicorp/hcl/v2

# Tidy dependencies
go mod tidy
```

## Performance Considerations

1. **Parser Reuse**: Create one `Parser` instance and reuse it
2. **Memory**: Large Terraform files are parsed in memory
3. **Concurrency**: Parser is not thread-safe; use separate instances per goroutine

## Error Handling

### Error Patterns

```go
// Return errors with context
if err != nil {
    return fmt.Errorf("failed to parse file %s: %w", filename, err)
}

// Collect multiple errors instead of failing fast
result.Errors = append(result.Errors, fmt.Sprintf("warning: %v", err))
```

### Custom Error Types

Consider adding custom error types for better error handling:

```go
type ParseError struct {
    File    string
    Line    int
    Message string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("%s:%d: %s", e.File, e.Line, e.Message)
}
```

## Debugging

### Logging

Add debug logging during development:

```go
import "log"

// In development
log.Printf("Parsing file: %s", filename)
log.Printf("Found %d variables", len(variables))
```

### Testing Individual Components

```go
// Test parser in isolation
func TestDebugParser(t *testing.T) {
    content := `variable "test" { type = string }`
    p := parser.NewParser()
    // ... test specific behavior
}
```

## Release Process

1. Update version in `go.mod`
2. Run all tests: `go test ./...`
3. Update CHANGELOG.md
4. Create git tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
5. Push tag: `git push origin v1.0.0`

## Future Enhancements

### Planned Features

1. **Complex Object Support**: Better handling of nested object types
2. **Advanced Validation**: Parse and convert complex validation expressions
3. **CLI Tool**: Command-line interface for standalone usage
4. **Additional Formats**: OpenAPI, GraphQL schema generation
5. **Performance**: Streaming parser for very large files

### Contributing

When adding features:
1. Write tests first (TDD)
2. Update documentation
3. Ensure backward compatibility
4. Add examples for new features

## Resources

- [HCL Specification](https://github.com/hashicorp/hcl/blob/main/hclsyntax/spec.md)
- [JSON Schema Draft 7](https://json-schema.org/draft-07/schema)
- [Terraform Variable Configuration](https://www.terraform.io/language/values/variables)
- [Go Project Layout](https://github.com/golang-standards/project-layout)

## Support & Community

- GitHub Issues: Bug reports and feature requests
- Discussions: Questions and general discussion
- Pull Requests: Code contributions welcome

## Best Practices

1. **Always write tests** for new features
2. **Keep packages focused** - single responsibility
3. **Document exported functions** with clear examples
4. **Handle errors explicitly** - no silent failures
5. **Use semantic versioning** for releases
6. **Maintain backward compatibility** when possible
