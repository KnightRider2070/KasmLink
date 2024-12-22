package tests

import (
	"encoding/json"
	embedfiles "kasmlink/embedded"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xeipuuv/gojsonschema"
	"kasmlink/pkg/dockercompose"
)

// Mock EmbedFile for the schema
var mockComposeSpec = []byte(`
{
  "$schema": "https://json-schema.org/draft/2019-09/schema#",
  "type": "object",
  "properties": {
    "version": { "type": "string" },
    "services": {
      "type": "object",
      "patternProperties": {
        "^[a-zA-Z0-9._-]+$": {
          "type": "object",
          "properties": {
            "image": { "type": "string" },
            "build": { 
              "type": "object",
              "properties": {
                "context": { "type": "string" },
                "dockerfile": { "type": "string" }
              },
              "additionalProperties": false
            }
          },
          "additionalProperties": false
        }
      },
      "additionalProperties": false
    }
  },
  "required": ["version", "services"],
  "additionalProperties": false
}
`)

// TestValidateDockerCompose tests the ValidateDockerCompose function.
func TestValidateDockerCompose(t *testing.T) {
	// Mock embedfiles.ComposeSpec
	embedfiles.ComposeSpec = mockComposeSpec

	t.Run("Positive Case - Valid DockerCompose Structure", func(t *testing.T) {
		validCompose := dockercompose.DockerCompose{
			Version: "3.9",
			Services: map[string]dockercompose.ServiceDefinition{
				"web": {
					Image: "nginx",
				},
			},
		}

		err := dockercompose.ValidateDockerCompose(validCompose)
		assert.NoError(t, err, "Expected no validation error for valid DockerCompose")
	})

	t.Run("Negative Case - Missing Required Field", func(t *testing.T) {
		invalidCompose := dockercompose.DockerCompose{
			Services: map[string]dockercompose.ServiceDefinition{
				"web": {
					Image: "nginx",
				},
			},
		}

		err := dockercompose.ValidateDockerCompose(invalidCompose)
		assert.Error(t, err, "Expected validation error for missing 'version'")
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("Negative Case - Invalid Field", func(t *testing.T) {
		// Use raw JSON to simulate an invalid field
		invalidCompose := map[string]interface{}{
			"version": "3.9",
			"services": map[string]interface{}{
				"web": map[string]interface{}{
					"image":        "nginx",
					"invalidField": "unexpected", // Invalid field not in schema
				},
			},
		}

		// Marshal the invalid structure into JSON
		jsonData, err := json.Marshal(invalidCompose)
		assert.NoError(t, err, "Failed to marshal invalid DockerCompose")

		// Load schema and validate
		schemaLoader := gojsonschema.NewBytesLoader(embedfiles.ComposeSpec)
		documentLoader := gojsonschema.NewStringLoader(string(jsonData))

		result, err := gojsonschema.Validate(schemaLoader, documentLoader)
		assert.NoError(t, err, "Schema validation encountered an unexpected error")

		// Ensure validation fails and contains the correct error message
		assert.False(t, result.Valid(), "Expected validation to fail for invalid field")
		if !result.Valid() {
			for _, desc := range result.Errors() {
				assert.Contains(t, desc.String(), "invalidField", "Error should reference the invalid field")
			}
		}
	})
}
