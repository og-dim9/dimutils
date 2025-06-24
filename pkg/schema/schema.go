package schema

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// SchemaType represents the type of a field in the schema
type SchemaType string

const (
	TypeString  SchemaType = "string"
	TypeNumber  SchemaType = "number"
	TypeInteger SchemaType = "integer"
	TypeBoolean SchemaType = "boolean"
	TypeArray   SchemaType = "array"
	TypeObject  SchemaType = "object"
	TypeNull    SchemaType = "null"
)

// Schema represents a JSON schema structure
type Schema struct {
	Type       SchemaType             `json:"type,omitempty"`
	Properties map[string]*Schema     `json:"properties,omitempty"`
	Items      *Schema                `json:"items,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Examples   []interface{}          `json:"examples,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Version    string                 `json:"version,omitempty"`
	Schema     string                 `json:"$schema,omitempty"`
	ID         string                 `json:"$id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// Config holds configuration for schema operations
type Config struct {
	Pretty          bool
	Evolve          bool
	ForceOptional   bool
	OutputFile      string
	InputFile       string
	SchemaFile      string
	RegistryPath    string
	Validate        bool
	Generate        bool
	Merge           bool
	Version         string
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Pretty:       true,
		RegistryPath: ".schema-registry",
		Version:      "1.0.0",
	}
}

// Run is the main entry point for schema functionality
func Run(args []string) error {
	config := DefaultConfig()
	
	// Parse arguments
	if err := parseArgs(args, &config); err != nil {
		return err
	}

	// Execute the requested operation
	switch {
	case config.Generate:
		return generateSchema(config)
	case config.Validate:
		return validateData(config)
	case config.Merge:
		return mergeSchemas(config)
	default:
		return printHelp()
	}
}

func parseArgs(args []string, config *Config) error {
	for i, arg := range args {
		switch arg {
		case "generate", "gen":
			config.Generate = true
		case "validate", "val":
			config.Validate = true
		case "merge":
			config.Merge = true
		case "--pretty":
			config.Pretty = true
		case "--no-pretty":
			config.Pretty = false
		case "--evolve":
			config.Evolve = true
		case "--force-optional":
			config.ForceOptional = true
		case "--output", "-o":
			if i+1 < len(args) {
				config.OutputFile = args[i+1]
			}
		case "--input", "-i":
			if i+1 < len(args) {
				config.InputFile = args[i+1]
			}
		case "--schema", "-s":
			if i+1 < len(args) {
				config.SchemaFile = args[i+1]
			}
		case "--registry", "-r":
			if i+1 < len(args) {
				config.RegistryPath = args[i+1]
			}
		case "--version", "-v":
			if i+1 < len(args) {
				config.Version = args[i+1]
			}
		case "-h", "--help":
			return printHelp()
		}
	}
	return nil
}

func printHelp() error {
	help := `Usage: schema <command> [options]

Commands:
  generate, gen     Generate schema from JSON data
  validate, val     Validate JSON data against schema
  merge            Merge multiple schemas

Options:
  --input, -i FILE      Input JSON file (default: stdin)
  --output, -o FILE     Output schema file (default: stdout)
  --schema, -s FILE     Schema file for validation
  --registry, -r PATH   Schema registry path (default: .schema-registry)
  --version, -v VER     Schema version (default: 1.0.0)
  --pretty             Pretty print output (default: true)
  --no-pretty         Disable pretty printing
  --evolve             Enable schema evolution
  --force-optional     Make all fields optional when merging
  -h, --help           Show this help message

Examples:
  cat data.json | schema generate --output schema.json
  schema validate --input data.json --schema schema.json
  schema merge --schema schema1.json schema2.json --output merged.json`

	fmt.Println(help)
	return nil
}

// generateSchema creates a schema from JSON data
func generateSchema(config Config) error {
	var input *os.File
	var err error

	// Set up input source
	if config.InputFile != "" {
		input, err = os.Open(config.InputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	// Process input data
	schema := &Schema{
		Schema:     "https://json-schema.org/draft/2020-12/schema",
		Type:       TypeObject,
		Properties: make(map[string]*Schema),
		Version:    config.Version,
	}

	scanner := bufio.NewScanner(input)
	recordCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			return fmt.Errorf("error parsing JSON on line %d: %w", recordCount+1, err)
		}

		if err := updateSchemaFromData(schema, data, config.Evolve); err != nil {
			return fmt.Errorf("error updating schema: %w", err)
		}

		recordCount++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	if recordCount == 0 {
		return fmt.Errorf("no valid JSON data found")
	}

	// Output the schema
	return outputSchema(schema, config)
}

// updateSchemaFromData updates the schema based on new data
func updateSchemaFromData(schema *Schema, data interface{}, evolve bool) error {
	switch v := data.(type) {
	case map[string]interface{}:
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}

		for key, value := range v {
			if existingSchema, exists := schema.Properties[key]; exists && evolve {
				// Evolve existing field schema
				if err := mergeFieldSchema(existingSchema, value); err != nil {
					return err
				}
			} else if !exists {
				// Add new field
				fieldSchema, err := inferSchemaFromValue(value)
				if err != nil {
					return err
				}
				schema.Properties[key] = fieldSchema
			}
		}

		// Update required fields (fields that appear in all records)
		if evolve {
			updateRequiredFields(schema, v)
		}

	default:
		return fmt.Errorf("top-level data must be an object")
	}

	return nil
}

// inferSchemaFromValue creates a schema for a single value
func inferSchemaFromValue(value interface{}) (*Schema, error) {
	schema := &Schema{}

	switch v := value.(type) {
	case nil:
		schema.Type = TypeNull
	case bool:
		schema.Type = TypeBoolean
	case float64:
		if v == float64(int64(v)) {
			schema.Type = TypeInteger
		} else {
			schema.Type = TypeNumber
		}
	case string:
		schema.Type = TypeString
	case []interface{}:
		schema.Type = TypeArray
		if len(v) > 0 {
			// Infer schema for array items
			itemSchema, err := inferSchemaFromValue(v[0])
			if err != nil {
				return nil, err
			}
			schema.Items = itemSchema
		}
	case map[string]interface{}:
		schema.Type = TypeObject
		schema.Properties = make(map[string]*Schema)
		for key, val := range v {
			fieldSchema, err := inferSchemaFromValue(val)
			if err != nil {
				return nil, err
			}
			schema.Properties[key] = fieldSchema
		}
	default:
		return nil, fmt.Errorf("unsupported type: %T", value)
	}

	return schema, nil
}

// mergeFieldSchema merges schema information from a new value
func mergeFieldSchema(schema *Schema, value interface{}) error {
	newSchema, err := inferSchemaFromValue(value)
	if err != nil {
		return err
	}

	// If types are different, evolve to more general type
	if schema.Type != newSchema.Type {
		schema.Type = getCompatibleType(schema.Type, newSchema.Type)
	}

	// Merge object properties
	if schema.Type == TypeObject && newSchema.Properties != nil {
		if schema.Properties == nil {
			schema.Properties = make(map[string]*Schema)
		}
		for key, newField := range newSchema.Properties {
			if existingField, exists := schema.Properties[key]; exists {
				mergeFieldSchema(existingField, value.(map[string]interface{})[key])
			} else {
				schema.Properties[key] = newField
			}
		}
	}

	return nil
}

// getCompatibleType returns a type that can accommodate both types
func getCompatibleType(type1, type2 SchemaType) SchemaType {
	if type1 == type2 {
		return type1
	}

	// Type promotion rules
	typeHierarchy := map[SchemaType]int{
		TypeNull:    0,
		TypeBoolean: 1,
		TypeInteger: 2,
		TypeNumber:  3,
		TypeString:  4,
		TypeArray:   5,
		TypeObject:  6,
	}

	// Promote to the more general type
	if typeHierarchy[type1] > typeHierarchy[type2] {
		return type1
	}
	return type2
}

// updateRequiredFields updates the required fields list
func updateRequiredFields(schema *Schema, data map[string]interface{}) {
	if schema.Required == nil {
		// First record - all fields are potentially required
		for key := range data {
			schema.Required = append(schema.Required, key)
		}
		sort.Strings(schema.Required)
		return
	}

	// Remove fields that are not in this record
	var newRequired []string
	for _, field := range schema.Required {
		if _, exists := data[field]; exists {
			newRequired = append(newRequired, field)
		}
	}
	schema.Required = newRequired
}

// outputSchema outputs the schema to the specified destination
func outputSchema(schema *Schema, config Config) error {
	var output []byte
	var err error

	if config.Pretty {
		output, err = json.MarshalIndent(schema, "", "  ")
	} else {
		output, err = json.Marshal(schema)
	}

	if err != nil {
		return fmt.Errorf("error marshaling schema: %w", err)
	}

	if config.OutputFile != "" {
		return os.WriteFile(config.OutputFile, output, 0644)
	}

	fmt.Println(string(output))
	return nil
}

// validateData validates JSON data against a schema
func validateData(config Config) error {
	if config.SchemaFile == "" {
		return fmt.Errorf("schema file required for validation")
	}

	// Load schema
	schemaData, err := os.ReadFile(config.SchemaFile)
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return fmt.Errorf("error parsing schema: %w", err)
	}

	// Set up input source
	var input *os.File
	if config.InputFile != "" {
		input, err = os.Open(config.InputFile)
		if err != nil {
			return fmt.Errorf("error opening input file: %w", err)
		}
		defer input.Close()
	} else {
		input = os.Stdin
	}

	// Validate each line of input
	scanner := bufio.NewScanner(input)
	lineNumber := 0
	validCount := 0
	errorCount := 0

	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: Invalid JSON: %v\n", lineNumber, err)
			errorCount++
			continue
		}

		if err := validateValue(data, &schema, ""); err != nil {
			fmt.Fprintf(os.Stderr, "Line %d: Validation error: %v\n", lineNumber, err)
			errorCount++
		} else {
			validCount++
		}
	}

	fmt.Printf("Validation complete: %d valid, %d errors\n", validCount, errorCount)

	if errorCount > 0 {
		return fmt.Errorf("validation failed with %d errors", errorCount)
	}

	return nil
}

// validateValue validates a value against a schema
func validateValue(value interface{}, schema *Schema, path string) error {
	// Type validation
	actualType := getValueType(value)
	if schema.Type != "" && actualType != schema.Type {
		return fmt.Errorf("type mismatch at %s: expected %s, got %s", path, schema.Type, actualType)
	}

	// Object validation
	if schema.Type == TypeObject && schema.Properties != nil {
		obj, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object at %s", path)
		}

		// Check required fields
		for _, required := range schema.Required {
			if _, exists := obj[required]; !exists {
				return fmt.Errorf("missing required field: %s%s", path, required)
			}
		}

		// Validate properties
		for key, val := range obj {
			if propSchema, exists := schema.Properties[key]; exists {
				newPath := path + "." + key
				if path == "" {
					newPath = key
				}
				if err := validateValue(val, propSchema, newPath); err != nil {
					return err
				}
			}
		}
	}

	// Array validation
	if schema.Type == TypeArray && schema.Items != nil {
		arr, ok := value.([]interface{})
		if !ok {
			return fmt.Errorf("expected array at %s", path)
		}

		for i, item := range arr {
			itemPath := fmt.Sprintf("%s[%d]", path, i)
			if err := validateValue(item, schema.Items, itemPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// getValueType returns the SchemaType for a value
func getValueType(value interface{}) SchemaType {
	switch v := value.(type) {
	case nil:
		return TypeNull
	case bool:
		return TypeBoolean
	case float64:
		if v == float64(int64(v)) {
			return TypeInteger
		}
		return TypeNumber
	case string:
		return TypeString
	case []interface{}:
		return TypeArray
	case map[string]interface{}:
		return TypeObject
	default:
		return TypeString // fallback
	}
}

// mergeSchemas merges multiple schemas
func mergeSchemas(config Config) error {
	if config.SchemaFile == "" {
		return fmt.Errorf("at least one schema file required for merging")
	}

	// Parse schema files from args (they come after the --schema flag)
	schemaFiles := []string{config.SchemaFile}
	
	// Load and merge all schemas
	var mergedSchema *Schema
	
	for i, schemaFile := range schemaFiles {
		schemaData, err := os.ReadFile(schemaFile)
		if err != nil {
			return fmt.Errorf("error reading schema file %s: %w", schemaFile, err)
		}

		var schema Schema
		if err := json.Unmarshal(schemaData, &schema); err != nil {
			return fmt.Errorf("error parsing schema %s: %w", schemaFile, err)
		}

		if i == 0 {
			mergedSchema = &schema
		} else {
			if err := mergeTwoSchemas(mergedSchema, &schema, config.ForceOptional); err != nil {
				return fmt.Errorf("error merging schema %s: %w", schemaFile, err)
			}
		}
	}

	if mergedSchema == nil {
		return fmt.Errorf("no schemas to merge")
	}

	// Update version
	mergedSchema.Version = config.Version

	// Output the merged schema
	return outputSchema(mergedSchema, config)
}

// mergeTwoSchemas merges two schemas together
func mergeTwoSchemas(target, source *Schema, forceOptional bool) error {
	// Merge types - use more general type if different
	if target.Type != source.Type {
		target.Type = getCompatibleType(target.Type, source.Type)
	}

	// Merge object properties
	if target.Type == TypeObject {
		if target.Properties == nil {
			target.Properties = make(map[string]*Schema)
		}

		// Add properties from source
		for key, sourceProp := range source.Properties {
			if targetProp, exists := target.Properties[key]; exists {
				// Merge existing property
				if err := mergeTwoSchemas(targetProp, sourceProp, forceOptional); err != nil {
					return err
				}
			} else {
				// Add new property
				target.Properties[key] = sourceProp
			}
		}

		// Merge required fields
		if forceOptional {
			// If force optional, clear required fields
			target.Required = nil
		} else {
			// Only keep fields that are required in both schemas
			target.Required = intersectStringSlices(target.Required, source.Required)
		}
	}

	// Merge array items
	if target.Type == TypeArray {
		if target.Items != nil && source.Items != nil {
			if err := mergeTwoSchemas(target.Items, source.Items, forceOptional); err != nil {
				return err
			}
		} else if source.Items != nil {
			target.Items = source.Items
		}
	}

	// Merge examples
	target.Examples = append(target.Examples, source.Examples...)

	// Update metadata
	if target.Metadata == nil {
		target.Metadata = make(map[string]interface{})
	}
	for key, value := range source.Metadata {
		target.Metadata[key] = value
	}

	return nil
}

// intersectStringSlices returns the intersection of two string slices
func intersectStringSlices(a, b []string) []string {
	setA := make(map[string]bool)
	for _, item := range a {
		setA[item] = true
	}

	var result []string
	for _, item := range b {
		if setA[item] {
			result = append(result, item)
		}
	}

	sort.Strings(result)
	return result
}