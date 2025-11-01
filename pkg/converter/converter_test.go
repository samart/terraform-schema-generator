package converter

import (
	"testing"

	"github.com/samart/terraform-schema-generator/pkg/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertToJSONSchema7(t *testing.T) {
	converter := NewConverter()

	t.Run("convert simple variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:        "instance_type",
					Type:        "string",
					Description: "EC2 instance type",
					Default:     "t2.micro",
					Required:    false,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema.Schema)
		assert.Equal(t, "object", schema.Type)
		assert.Len(t, schema.Properties, 1)
		assert.Equal(t, "string", schema.Properties["instance_type"].Type)
		assert.Empty(t, schema.Required)
	})

	t.Run("convert required variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:     "vpc_id",
					Type:     "string",
					Required: true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Contains(t, schema.Required, "vpc_id")
	})

	t.Run("convert sensitive variable", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name:      "db_password",
					Type:      "string",
					Sensitive: true,
					Required:  true,
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.True(t, schema.Properties["db_password"].WriteOnly)
	})

	t.Run("convert number type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "instance_count",
					Type: "number",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "number", schema.Properties["instance_count"].Type)
	})

	t.Run("convert boolean type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "enable_monitoring",
					Type: "bool",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "boolean", schema.Properties["enable_monitoring"].Type)
	})

	t.Run("convert list type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "availability_zones",
					Type: "list(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "array", schema.Properties["availability_zones"].Type)
	})

	t.Run("convert map type", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{
				{
					Name: "tags",
					Type: "map(string)",
				},
			},
		}

		schema, err := converter.ConvertToJSONSchema7(parseResult)
		require.NoError(t, err)
		assert.Equal(t, "object", schema.Properties["tags"].Type)
	})

	t.Run("error on empty variables", func(t *testing.T) {
		parseResult := &parser.ParseResult{
			Variables: []parser.Variable{},
		}

		_, err := converter.ConvertToJSONSchema7(parseResult)
		assert.Error(t, err)
	})
}
