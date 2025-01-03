package dockercompose_tests

import (
	"encoding/json"
	embedfiles "kasmlink/embedded"
	"kasmlink/pkg/dockercompose"
	"reflect"
	"strings"
	"testing"

	"github.com/xeipuuv/gojsonschema"
)

// GenerateTestInstance creates a JSON representation of a struct with example values for all fields.
func GenerateTestInstance(structType interface{}) (string, error) {
	v := reflect.ValueOf(structType)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", nil
	}

	// Create a map to hold the test data
	data := map[string]interface{}{}

	// Iterate through the fields
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Tag.Get("json")

		// Skip fields without JSON tags or marked with "-"
		if fieldName == "" || fieldName == "-" {
			continue
		}

		// Remove ",omitempty" if present in the fieldName
		fieldName = strings.Split(fieldName, ",")[0]

		// Populate fields with schema-compliant example data
		switch fieldName {
		case "version":
			data[fieldName] = "3.9" // Example version
		case "services":
			data[fieldName] = map[string]interface{}{
				"web": map[string]interface{}{
					"image": "nginx",
					"ports": []map[string]interface{}{
						{"target": 80, "published": 8080, "protocol": "tcp", "mode": "host"},
					},
					"build": map[string]interface{}{"context": "./app", "dockerfile": "Dockerfile"},
					"environment": map[string]string{
						"APP_ENV": "production",
						"DEBUG":   "false",
					},
				},
				"db": map[string]interface{}{
					"image": "mysql",
					"environment": map[string]string{
						"MYSQL_ROOT_PASSWORD": "root",
						"MYSQL_DATABASE":      "test",
					},
				},
			}
		case "networks":
			data[fieldName] = map[string]interface{}{
				"default": map[string]interface{}{
					"driver":     "bridge",
					"attachable": true,
				},
				"custom_network": map[string]interface{}{
					"driver": "overlay",
				},
			}
		case "volumes":
			data[fieldName] = map[string]interface{}{
				"myvolume":      map[string]interface{}{"driver": "local"},
				"shared_volume": map[string]interface{}{"external": true},
			}
		case "configs":
			data[fieldName] = map[string]interface{}{
				"app_config": map[string]interface{}{"file": "config.json"},
				"shared_config": map[string]interface{}{
					"external": true,
				},
			}
		case "secrets":
			data[fieldName] = map[string]interface{}{
				"mysecret": map[string]interface{}{"file": "secret.txt"},
			}
		case "include":
			data[fieldName] = []string{"compose.override.yml", "compose.prod.yml"}
		case "name":
			data[fieldName] = "example_project"
		default:
			// Populate with default example values for other fields
			data[fieldName] = generateExampleValue(field.Type)
		}
	}

	// Convert to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// generateExampleValue creates example data based on the field type.
func generateExampleValue(t reflect.Type) interface{} {
	switch t.Kind() {
	case reflect.String:
		return "example"
	case reflect.Int, reflect.Int64:
		return 123
	case reflect.Bool:
		return true
	case reflect.Slice:
		// Handle slices of structs
		if t.Elem().Kind() == reflect.Struct {
			return []interface{}{generateExampleValue(t.Elem())}
		}
		return []interface{}{"example1", "example2"}
	case reflect.Map:
		// Handle maps with string keys and interface values
		return map[string]interface{}{"key": "value"}
	case reflect.Struct:
		// Recursively generate example values for nested structs
		instance := map[string]interface{}{}
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldName := strings.Split(field.Tag.Get("json"), ",")[0]
			if fieldName == "" || fieldName == "-" {
				continue
			}
			instance[fieldName] = generateExampleValue(field.Type)
		}
		return instance
	default:
		return nil
	}
}

// TestStructAgainstSchema validates the generated JSON against the provided schema.
func TestStructAgainstSchema(t *testing.T) {
	// Use the schema loader with the embedded schema
	schemaLoader := gojsonschema.NewBytesLoader(embedfiles.ComposeSpec)

	// Generate JSON from the DockerCompose struct
	instance, err := GenerateTestInstance(dockercompose.DockerCompose{})
	if err != nil {
		t.Fatalf("Failed to generate test instance: %v", err)
	}

	// Log the generated instance for debugging
	t.Logf("Generated JSON: %s", instance)

	// Create a loader for the generated JSON
	documentLoader := gojsonschema.NewStringLoader(instance)

	// Validate the JSON against the schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}

	// Check for validation errors
	if !result.Valid() {
		for _, err := range result.Errors() {
			t.Errorf("Validation error: %s\nInstance: %s", err, instance)
		}
		t.FailNow()
	}
}

// TestInvalidStructAgainstSchema validates invalid JSON against the schema to ensure errors are detected.
func TestInvalidStructAgainstSchema(t *testing.T) {
	// Use the schema loader with the embedded schema
	schemaLoader := gojsonschema.NewBytesLoader(embedfiles.ComposeSpec)

	// Create intentionally invalid JSON
	invalidInstance := `{
		"version": "3.9",
		"services": {
			"web": {
				"invalid_field": "invalid_value"
			}
		}
	}`

	// Log the invalid instance for debugging
	t.Logf("Invalid JSON: %s", invalidInstance)

	// Create a loader for the invalid JSON
	documentLoader := gojsonschema.NewStringLoader(invalidInstance)

	// Validate the JSON against the schema
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		t.Fatalf("Schema validation failed: %v", err)
	}

	// Ensure validation errors are detected
	if result.Valid() {
		t.Fatalf("Expected validation errors, but none were found")
	}

	// Log validation errors
	for _, err := range result.Errors() {
		t.Logf("Expected validation error: %s", err)
	}
}
