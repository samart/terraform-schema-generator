package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/samart/terraform-schema-generator/pkg/generator"
)

const version = "1.0.0"

var (
	// Global flags
	inputDir   string
	inputFile  string
	outputFile string
	validate   bool
	verbose    bool
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		// Error is already handled by cobra, just exit
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "terraform-schema-generator",
	Short: "Generate JSON Schema from Terraform configurations",
	Long: `terraform-schema-generator is a CLI tool that parses Terraform variable
definitions and generates JSON Schema Draft 7 specifications.

It supports reading from individual files or entire directories containing
Terraform configurations.`,
	Version:      version,
	RunE:         runGenerate,
	SilenceUsage: true, // Don't show usage on errors
}

func init() {
	// Input flags
	rootCmd.Flags().StringVarP(&inputDir, "dir", "d", "", "Directory containing Terraform files")
	rootCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Single Terraform file to process")

	// Output flags
	rootCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file path (default: stdout)")

	// Validation flags
	rootCmd.Flags().BoolVar(&validate, "validate", true, "Validate generated schema against JSON Schema Draft 7")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// Mark at least one input source as required
	rootCmd.MarkFlagsMutuallyExclusive("dir", "file")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Validate input
	if err := validateInput(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Generate schema
	schemaJSON, err := generateSchema()
	if err != nil {
		return err // Error already has context from generateSchema
	}

	// Write output
	if err := writeOutput(schemaJSON); err != nil {
		return err // Error already has context from writeOutput
	}

	// Success message
	if verbose {
		if outputFile != "" {
			fmt.Fprintf(os.Stderr, "✓ Schema successfully written to: %s\n", outputFile)
		} else {
			fmt.Fprintln(os.Stderr, "✓ Schema successfully generated")
		}
	}

	return nil
}

func validateInput() error {
	if inputDir == "" && inputFile == "" {
		return fmt.Errorf("either --dir or --file must be specified")
	}

	if inputDir != "" && inputFile != "" {
		return fmt.Errorf("only one of --dir or --file can be specified")
	}

	// Validate input exists
	if inputDir != "" {
		info, err := os.Stat(inputDir)
		if err != nil {
			return fmt.Errorf("cannot access directory: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("path is not a directory: %s", inputDir)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "→ Reading Terraform files from directory: %s\n", inputDir)
		}
	}

	if inputFile != "" {
		info, err := os.Stat(inputFile)
		if err != nil {
			return fmt.Errorf("cannot access file: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("path is a directory, not a file: %s", inputFile)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "→ Reading Terraform file: %s\n", inputFile)
		}
	}

	return nil
}

func generateSchema() ([]byte, error) {
	if verbose {
		fmt.Fprintln(os.Stderr, "→ Parsing Terraform configuration...")
	}

	// Build generator
	gen := generator.New()

	// Add input source
	if inputDir != "" {
		gen = gen.FromDirectory(inputDir)
	} else {
		gen = gen.FromFile(inputFile)
	}

	// Parse
	gen = gen.Parse()

	// Check parsing errors
	if err := gen.Error(); err != nil {
		return nil, fmt.Errorf("parsing failed: %w", err)
	}

	// Show parse results if verbose
	if verbose {
		result, err := gen.ParseResult()
		if err != nil {
			return nil, fmt.Errorf("failed to get parse result: %w", err)
		}
		fmt.Fprintf(os.Stderr, "→ Found %d variables\n", len(result.Variables))
		fmt.Fprintln(os.Stderr, "→ Converting to JSON Schema Draft 7...")
	}

	// Convert to JSON Schema
	gen = gen.Convert()

	// Check conversion errors
	if err := gen.Error(); err != nil {
		return nil, fmt.Errorf("conversion failed: %w", err)
	}

	// Validate if requested
	if validate {
		if verbose {
			fmt.Fprintln(os.Stderr, "→ Validating against JSON Schema Draft 7 meta-schema...")
		}
		gen = gen.ValidateAgainstMetaSchema()

		// Check validation errors
		if err := gen.Error(); err != nil {
			return nil, fmt.Errorf("schema validation failed: %w", err)
		}

		if verbose {
			fmt.Fprintln(os.Stderr, "✓ Schema validation passed")
		}
	}

	// Get JSON bytes
	jsonBytes, err := gen.JSON()
	if err != nil {
		return nil, fmt.Errorf("failed to generate JSON: %w", err)
	}

	return jsonBytes, nil
}

func writeOutput(jsonBytes []byte) error {
	if outputFile != "" {
		// Write to file
		if verbose {
			fmt.Fprintf(os.Stderr, "→ Writing schema to file: %s\n", outputFile)
		}

		if err := os.WriteFile(outputFile, jsonBytes, 0644); err != nil {
			return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
		}
	} else {
		// Write to stdout
		if _, err := os.Stdout.Write(jsonBytes); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}
