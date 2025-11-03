package generator

import (
	"fmt"
	"io"
	"os"

	"github.com/samart/terraform-schema-generator/pkg/converter"
	"github.com/samart/terraform-schema-generator/pkg/parser"
	"github.com/samart/terraform-schema-generator/pkg/validator"
)

// Generator provides a fluent API for generating JSON schemas from Terraform files
type Generator struct {
	files      map[string]io.Reader
	result     *parser.ParseResult
	schema     *converter.JSONSchema7
	schemaJSON []byte
	errors     []error
}

// New creates a new Generator instance
func New() *Generator {
	return &Generator{
		files:  make(map[string]io.Reader),
		errors: []error{},
	}
}

// FromFile adds a Terraform file from the filesystem
func (g *Generator) FromFile(path string) *Generator {
	file, err := os.Open(path)
	if err != nil {
		g.errors = append(g.errors, fmt.Errorf("failed to open file %s: %w", path, err))
		return g
	}
	g.files[path] = file
	return g
}

// FromString adds Terraform configuration from a string
func (g *Generator) FromString(name, content string) *Generator {
	g.files[name] = &stringReader{content: content, pos: 0}
	return g
}

// FromReader adds Terraform configuration from an io.Reader
func (g *Generator) FromReader(name string, reader io.Reader) *Generator {
	g.files[name] = reader
	return g
}

// FromDirectory adds all .tf files from a directory
func (g *Generator) FromDirectory(path string) *Generator {
	entries, err := os.ReadDir(path)
	if err != nil {
		g.errors = append(g.errors, fmt.Errorf("failed to read directory %s: %w", path, err))
		return g
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if len(name) < 3 || name[len(name)-3:] != ".tf" {
			continue
		}

		filePath := path + "/" + name
		g.FromFile(filePath)
	}

	return g
}

// Parse parses all added Terraform files
func (g *Generator) Parse() *Generator {
	if len(g.errors) > 0 {
		return g
	}

	p := parser.NewParser()
	result, err := p.ParseFiles(g.files)
	if err != nil {
		g.errors = append(g.errors, fmt.Errorf("parse failed: %w", err))
		return g
	}

	g.result = result
	return g
}

// Convert converts the parsed result to JSON Schema
func (g *Generator) Convert() *Generator {
	if len(g.errors) > 0 {
		return g
	}

	if g.result == nil {
		g.errors = append(g.errors, fmt.Errorf("no parse result available, call Parse() first"))
		return g
	}

	c := converter.NewConverter()
	schema, err := c.ConvertToJSONSchema7(g.result)
	if err != nil {
		g.errors = append(g.errors, fmt.Errorf("conversion failed: %w", err))
		return g
	}

	g.schema = schema

	// Convert to JSON
	schemaJSON, err := c.ToJSON(schema)
	if err != nil {
		g.errors = append(g.errors, fmt.Errorf("JSON conversion failed: %w", err))
		return g
	}

	g.schemaJSON = schemaJSON
	return g
}

// Validate validates the generated schema
func (g *Generator) Validate() *Generator {
	if len(g.errors) > 0 {
		return g
	}

	if g.schema == nil {
		g.errors = append(g.errors, fmt.Errorf("no schema available, call Convert() first"))
		return g
	}

	v := validator.NewValidator()
	if err := v.ValidateSchema(g.schema); err != nil {
		g.errors = append(g.errors, fmt.Errorf("schema validation failed: %w", err))
		return g
	}

	return g
}

// ValidateAgainstMetaSchema validates against JSON Schema Draft 7 meta-schema
func (g *Generator) ValidateAgainstMetaSchema() *Generator {
	if len(g.errors) > 0 {
		return g
	}

	if g.schemaJSON == nil {
		g.errors = append(g.errors, fmt.Errorf("no schema JSON available, call Convert() first"))
		return g
	}

	v := validator.NewMetaSchemaValidator()
	if err := v.ValidateAgainstMetaSchema(g.schemaJSON); err != nil {
		g.errors = append(g.errors, fmt.Errorf("meta-schema validation failed: %w", err))
		return g
	}

	return g
}

// ToFile writes the schema to a file
func (g *Generator) ToFile(path string) *Generator {
	if len(g.errors) > 0 {
		return g
	}

	if g.schemaJSON == nil {
		g.errors = append(g.errors, fmt.Errorf("no schema JSON available, call Convert() first"))
		return g
	}

	if err := os.WriteFile(path, g.schemaJSON, 0644); err != nil {
		g.errors = append(g.errors, fmt.Errorf("failed to write to file %s: %w", path, err))
		return g
	}

	return g
}

// JSON returns the generated JSON schema as bytes
func (g *Generator) JSON() ([]byte, error) {
	if len(g.errors) > 0 {
		return nil, g.errors[0]
	}

	if g.schemaJSON == nil {
		return nil, fmt.Errorf("no schema JSON available, call Convert() first")
	}

	return g.schemaJSON, nil
}

// Schema returns the generated JSON Schema struct
func (g *Generator) Schema() (*converter.JSONSchema7, error) {
	if len(g.errors) > 0 {
		return nil, g.errors[0]
	}

	if g.schema == nil {
		return nil, fmt.Errorf("no schema available, call Convert() first")
	}

	return g.schema, nil
}

// ParseResult returns the parser result
func (g *Generator) ParseResult() (*parser.ParseResult, error) {
	if len(g.errors) > 0 {
		return nil, g.errors[0]
	}

	if g.result == nil {
		return nil, fmt.Errorf("no parse result available, call Parse() first")
	}

	return g.result, nil
}

// Error returns the first error encountered, or nil if no errors
func (g *Generator) Error() error {
	if len(g.errors) > 0 {
		return g.errors[0]
	}
	return nil
}

// Errors returns all errors encountered
func (g *Generator) Errors() []error {
	return g.errors
}

// stringReader implements io.Reader for strings
type stringReader struct {
	content string
	pos     int
}

func (sr *stringReader) Read(p []byte) (n int, err error) {
	if sr.pos >= len(sr.content) {
		return 0, io.EOF
	}
	n = copy(p, sr.content[sr.pos:])
	sr.pos += n
	return n, nil
}

// Quick helper functions for common use cases

// FromFileQuick generates a schema from a single file
func FromFileQuick(path string) ([]byte, error) {
	return New().FromFile(path).Parse().Convert().Validate().JSON()
}

// FromStringQuick generates a schema from a string
func FromStringQuick(content string) ([]byte, error) {
	return New().FromString("terraform.tf", content).Parse().Convert().Validate().JSON()
}

// FromDirectoryQuick generates a schema from all .tf files in a directory
func FromDirectoryQuick(path string) ([]byte, error) {
	return New().FromDirectory(path).Parse().Convert().Validate().JSON()
}
