package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// Variable represents a Terraform variable definition
type Variable struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Default     interface{}            `json:"default,omitempty"`
	Required    bool                   `json:"required"`
	Sensitive   bool                   `json:"sensitive,omitempty"`
	Validation  []ValidationRule       `json:"validation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationRule represents a Terraform variable validation block
type ValidationRule struct {
	Condition    string `json:"condition"`
	ErrorMessage string `json:"error_message"`
}

// Output represents a Terraform output definition
type Output struct {
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	Value       string      `json:"value,omitempty"`
	Sensitive   bool        `json:"sensitive,omitempty"`
	DependsOn   []string    `json:"depends_on,omitempty"`
}

// Provider represents a Terraform provider requirement
type Provider struct {
	Name    string `json:"name"`
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
}

// Resource represents a Terraform resource
type Resource struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Count int    `json:"count,omitempty"`
}

// Module represents a module dependency
type Module struct {
	Name    string `json:"name"`
	Source  string `json:"source"`
	Version string `json:"version,omitempty"`
}

// ParseResult contains the parsed Terraform variables
type ParseResult struct {
	Variables          []Variable `json:"variables"`
	Outputs            []Output   `json:"outputs,omitempty"`
	Providers          []Provider `json:"providers,omitempty"`
	Resources          []Resource `json:"resources,omitempty"`
	Modules            []Module   `json:"modules,omitempty"`
	TerraformVersion   string     `json:"terraform_version,omitempty"`
	Errors             []string   `json:"errors,omitempty"`
}

// Parser handles parsing of Terraform files
type Parser struct {
	parser *hclparse.Parser
}

// NewParser creates a new Terraform parser instance
func NewParser() *Parser {
	return &Parser{
		parser: hclparse.NewParser(),
	}
}

// ParseFiles parses multiple Terraform files and extracts all components
func (p *Parser) ParseFiles(files map[string]io.Reader) (*ParseResult, error) {
	result := &ParseResult{
		Variables: []Variable{},
		Outputs:   []Output{},
		Providers: []Provider{},
		Resources: []Resource{},
		Modules:   []Module{},
		Errors:    []string{},
	}

	for filename, reader := range files {
		// Read file content
		content, err := io.ReadAll(reader)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to read %s: %v", filename, err))
			continue
		}

		// Parse HCL file
		file, diags := p.parser.ParseHCL(content, filename)
		if diags.HasErrors() {
			result.Errors = append(result.Errors, fmt.Sprintf("Parse errors in %s: %s", filename, diags.Error()))
			continue
		}

		// Extract all components from the parsed file
		vars, err := p.extractVariables(file)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to extract variables from %s: %v", filename, err))
		} else {
			result.Variables = append(result.Variables, vars...)
		}

		outputs := p.extractOutputs(file)
		result.Outputs = append(result.Outputs, outputs...)

		resources := p.extractResources(file)
		result.Resources = append(result.Resources, resources...)

		modules := p.extractModules(file)
		result.Modules = append(result.Modules, modules...)

		// Extract terraform block for version and providers
		if filename == "versions.tf" || strings.Contains(string(content), "terraform {") {
			tfVersion, providers := p.extractTerraformBlock(file)
			if tfVersion != "" {
				result.TerraformVersion = tfVersion
			}
			result.Providers = append(result.Providers, providers...)
		}
	}

	return result, nil
}

// extractVariables extracts variable blocks from an HCL file
func (p *Parser) extractVariables(file *hcl.File) ([]Variable, error) {
	variables := []Variable{}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return variables, fmt.Errorf("unexpected body type")
	}

	for _, block := range body.Blocks {
		if block.Type != "variable" {
			continue
		}

		if len(block.Labels) == 0 {
			continue
		}

		variable := Variable{
			Name:     block.Labels[0],
			Required: true, // Default to required unless default is set
			Metadata: make(map[string]interface{}),
		}

		// Extract variable attributes
		if typeAttr, exists := block.Body.Attributes["type"]; exists {
			variable.Type = p.extractTypeString(typeAttr.Expr)
		}

		if descAttr, exists := block.Body.Attributes["description"]; exists {
			if val, diags := descAttr.Expr.Value(nil); !diags.HasErrors() {
				variable.Description = val.AsString()
			}
		}

		if defaultAttr, exists := block.Body.Attributes["default"]; exists {
			variable.Required = false
			if val, diags := defaultAttr.Expr.Value(nil); !diags.HasErrors() {
				variable.Default = p.convertCtyValue(val)
			}
		}

		if sensitiveAttr, exists := block.Body.Attributes["sensitive"]; exists {
			if val, diags := sensitiveAttr.Expr.Value(nil); !diags.HasErrors() {
				variable.Sensitive = val.True()
			}
		}

		// Extract validation blocks
		for _, validationBlock := range block.Body.Blocks {
			if validationBlock.Type == "validation" {
				rule := ValidationRule{}

				if condAttr, exists := validationBlock.Body.Attributes["condition"]; exists {
					rule.Condition = string(condAttr.Expr.Range().SliceBytes(file.Bytes))
				}

				if errMsgAttr, exists := validationBlock.Body.Attributes["error_message"]; exists {
					if val, diags := errMsgAttr.Expr.Value(nil); !diags.HasErrors() {
						rule.ErrorMessage = val.AsString()
					}
				}

				variable.Validation = append(variable.Validation, rule)
			}
		}

		variables = append(variables, variable)
	}

	return variables, nil
}

// extractTypeString extracts the type as a string from an expression
func (p *Parser) extractTypeString(expr hclsyntax.Expression) string {
	// Handle simple type references like string, number, bool
	if traversal, ok := expr.(*hclsyntax.ScopeTraversalExpr); ok {
		parts := []string{}
		for _, part := range traversal.Traversal {
			if name, ok := part.(hcl.TraverseRoot); ok {
				parts = append(parts, name.Name)
			}
		}
		return strings.Join(parts, ".")
	}

	// Handle complex types like list(string), map(number), object({...})
	if funcCall, ok := expr.(*hclsyntax.FunctionCallExpr); ok {
		return funcCall.Name
	}

	// Fallback: return the raw syntax
	return strings.TrimSpace(string(expr.Range().SliceBytes([]byte{})))
}

// convertCtyValue converts a cty.Value to a native Go type
func (p *Parser) convertCtyValue(val interface{}) interface{} {
	// This is a simplified conversion - in production you'd handle all cty types
	// For now, return as-is and let JSON marshaling handle it
	return val
}

// extractOutputs extracts output blocks from an HCL file
func (p *Parser) extractOutputs(file *hcl.File) []Output {
	outputs := []Output{}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return outputs
	}

	for _, block := range body.Blocks {
		if block.Type != "output" {
			continue
		}

		if len(block.Labels) == 0 {
			continue
		}

		output := Output{
			Name: block.Labels[0],
		}

		if descAttr, exists := block.Body.Attributes["description"]; exists {
			if val, diags := descAttr.Expr.Value(nil); !diags.HasErrors() {
				output.Description = val.AsString()
			}
		}

		if valAttr, exists := block.Body.Attributes["value"]; exists {
			output.Value = strings.TrimSpace(string(valAttr.Expr.Range().SliceBytes(file.Bytes)))
		}

		if sensitiveAttr, exists := block.Body.Attributes["sensitive"]; exists {
			if val, diags := sensitiveAttr.Expr.Value(nil); !diags.HasErrors() {
				output.Sensitive = val.True()
			}
		}

		outputs = append(outputs, output)
	}

	return outputs
}

// extractResources extracts resource blocks from an HCL file
func (p *Parser) extractResources(file *hcl.File) []Resource {
	resources := []Resource{}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return resources
	}

	for _, block := range body.Blocks {
		if block.Type != "resource" {
			continue
		}

		if len(block.Labels) < 2 {
			continue
		}

		resource := Resource{
			Type: block.Labels[0],
			Name: block.Labels[1],
		}

		// Check if resource uses count
		if countAttr, exists := block.Body.Attributes["count"]; exists {
			// Try to extract count value
			if val, diags := countAttr.Expr.Value(nil); !diags.HasErrors() {
				if val.Type().IsPrimitiveType() {
					resource.Count = 1 // Mark as using count
				}
			} else {
				resource.Count = 1 // Count expression exists
			}
		}

		resources = append(resources, resource)
	}

	return resources
}

// extractModules extracts module blocks from an HCL file
func (p *Parser) extractModules(file *hcl.File) []Module {
	modules := []Module{}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return modules
	}

	for _, block := range body.Blocks {
		if block.Type != "module" {
			continue
		}

		if len(block.Labels) == 0 {
			continue
		}

		module := Module{
			Name: block.Labels[0],
		}

		if sourceAttr, exists := block.Body.Attributes["source"]; exists {
			if val, diags := sourceAttr.Expr.Value(nil); !diags.HasErrors() {
				module.Source = val.AsString()
			}
		}

		if versionAttr, exists := block.Body.Attributes["version"]; exists {
			if val, diags := versionAttr.Expr.Value(nil); !diags.HasErrors() {
				module.Version = val.AsString()
			}
		}

		modules = append(modules, module)
	}

	return modules
}

// extractTerraformBlock extracts terraform version and provider requirements
func (p *Parser) extractTerraformBlock(file *hcl.File) (string, []Provider) {
	var terraformVersion string
	providers := []Provider{}

	body, ok := file.Body.(*hclsyntax.Body)
	if !ok {
		return terraformVersion, providers
	}

	for _, block := range body.Blocks {
		if block.Type != "terraform" {
			continue
		}

		// Extract required_version
		if versionAttr, exists := block.Body.Attributes["required_version"]; exists {
			if val, diags := versionAttr.Expr.Value(nil); !diags.HasErrors() {
				terraformVersion = val.AsString()
			}
		}

		// Extract required_providers block
		for _, nestedBlock := range block.Body.Blocks {
			if nestedBlock.Type != "required_providers" {
				continue
			}

			// Each attribute in required_providers is a provider
			for name, attr := range nestedBlock.Body.Attributes {
				provider := Provider{
					Name: name,
				}

				// Parse provider configuration (object with source and version)
				if val, diags := attr.Expr.Value(nil); !diags.HasErrors() {
					if val.Type().IsObjectType() {
						// Extract source
						if sourceVal := val.GetAttr("source"); !sourceVal.IsNull() {
							provider.Source = sourceVal.AsString()
						}
						// Extract version
						if versionVal := val.GetAttr("version"); !versionVal.IsNull() {
							provider.Version = versionVal.AsString()
						}
					}
				}

				providers = append(providers, provider)
			}
		}
	}

	return terraformVersion, providers
}
