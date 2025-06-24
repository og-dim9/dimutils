package validate

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	Valid      bool              `json:"valid"`
	Errors     []ValidationError `json:"errors,omitempty"`
	Warnings   []ValidationError `json:"warnings,omitempty"`
	Statistics ValidationStats   `json:"statistics"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Path        string `json:"path"`
	Field       string `json:"field,omitempty"`
	Value       string `json:"value,omitempty"`
	Message     string `json:"message"`
	Rule        string `json:"rule"`
	Severity    string `json:"severity"` // error, warning, info
	LineNumber  int    `json:"line_number,omitempty"`
	ColumnIndex int    `json:"column_index,omitempty"`
}

// ValidationStats contains validation statistics
type ValidationStats struct {
	TotalRecords    int `json:"total_records"`
	ValidRecords    int `json:"valid_records"`
	InvalidRecords  int `json:"invalid_records"`
	ErrorCount      int `json:"error_count"`
	WarningCount    int `json:"warning_count"`
	ProcessingTime  string `json:"processing_time"`
}

// ValidationConfig holds validation configuration
type ValidationConfig struct {
	SchemaFile      string            `json:"schema_file,omitempty"`
	SchemaContent   string            `json:"schema_content,omitempty"`
	Rules           []ValidationRule  `json:"rules,omitempty"`
	InputFile       string            `json:"input_file,omitempty"`
	OutputFile      string            `json:"output_file,omitempty"`
	Format          string            `json:"format"` // json, csv, text
	StrictMode      bool              `json:"strict_mode"`
	MaxErrors       int               `json:"max_errors"`
	IgnoreFields    []string          `json:"ignore_fields,omitempty"`
	RequiredFields  []string          `json:"required_fields,omitempty"`
	CustomRules     map[string]string `json:"custom_rules,omitempty"`
	Verbose         bool              `json:"verbose"`
	ShowWarnings    bool              `json:"show_warnings"`
}

// ValidationRule represents a custom validation rule
type ValidationRule struct {
	Field       string `json:"field"`
	Type        string `json:"type"`        // string, number, boolean, email, url, regex, range, length
	Required    bool   `json:"required"`
	Pattern     string `json:"pattern,omitempty"`
	MinLength   int    `json:"min_length,omitempty"`
	MaxLength   int    `json:"max_length,omitempty"`
	MinValue    *float64 `json:"min_value,omitempty"`
	MaxValue    *float64 `json:"max_value,omitempty"`
	AllowedValues []string `json:"allowed_values,omitempty"`
	Message     string `json:"message,omitempty"`
}

// Schema represents a simple JSON schema
type Schema struct {
	Type       string                 `json:"type"`
	Properties map[string]SchemaField `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

// SchemaField represents a field in a schema
type SchemaField struct {
	Type        string   `json:"type"`
	Format      string   `json:"format,omitempty"`
	Pattern     string   `json:"pattern,omitempty"`
	MinLength   *int     `json:"minLength,omitempty"`
	MaxLength   *int     `json:"maxLength,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Enum        []string `json:"enum,omitempty"`
	Description string   `json:"description,omitempty"`
}

// Run is the main entry point for validate functionality
func Run(args []string) error {
	if len(args) == 0 {
		return printHelp()
	}

	command := args[0]
	switch command {
	case "json":
		return validateJSON(args[1:])
	case "schema":
		return validateWithSchema(args[1:])
	case "rules":
		return validateWithRules(args[1:])
	case "generate-schema":
		return generateSchema(args[1:])
	case "create-rules":
		return createRules(args[1:])
	case "-h", "--help":
		return printHelp()
	default:
		// Default to JSON validation if first arg is a file
		return validateJSON(args)
	}
}

func printHelp() error {
	help := `Usage: validate <command> [options]

Data validation and schema checking tool.

Commands:
  json <file>                     Validate JSON format and structure
  schema <schema> <data>          Validate data against JSON schema
  rules <rules> <data>            Validate data against custom rules
  generate-schema <data>          Generate JSON schema from data
  create-rules                    Interactively create validation rules

Options:
  --format FORMAT                 Output format: json, text, csv (default: text)
  --output, -o FILE               Output file for results
  --strict                        Strict validation mode
  --max-errors N                  Maximum number of errors to report (default: 100)
  --ignore-fields FIELDS          Comma-separated list of fields to ignore
  --required-fields FIELDS        Comma-separated list of required fields
  --verbose, -v                   Verbose output
  --show-warnings                 Show warnings in addition to errors
  -h, --help                      Show this help message

Examples:
  validate json data.json
  validate schema schema.json data.json
  validate rules validation-rules.json data.json
  validate generate-schema sample-data.json > schema.json
  validate json --format json --output results.json data.json`

	fmt.Println(help)
	return nil
}

func validateJSON(args []string) error {
	config := ValidationConfig{
		Format:       "text",
		MaxErrors:    100,
		ShowWarnings: true,
	}

	// Parse arguments
	var inputFile string
	for i, arg := range args {
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				config.Format = args[i+1]
			}
		case "--output", "-o":
			if i+1 < len(args) {
				config.OutputFile = args[i+1]
			}
		case "--strict":
			config.StrictMode = true
		case "--max-errors":
			if i+1 < len(args) {
				if maxErrors, err := strconv.Atoi(args[i+1]); err == nil {
					config.MaxErrors = maxErrors
				}
			}
		case "--verbose", "-v":
			config.Verbose = true
		case "--show-warnings":
			config.ShowWarnings = true
		default:
			if !strings.HasPrefix(arg, "-") && inputFile == "" {
				inputFile = arg
			}
		}
	}

	if inputFile == "" {
		return fmt.Errorf("input file required")
	}

	config.InputFile = inputFile
	return performValidation(config)
}

func validateWithSchema(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("schema file and data file required")
	}

	config := ValidationConfig{
		SchemaFile: args[0],
		InputFile:  args[1],
		Format:     "text",
		MaxErrors:  100,
	}

	// Parse additional arguments
	for i := 2; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--format", "-f":
			if i+1 < len(args) {
				config.Format = args[i+1]
				i++
			}
		case "--output", "-o":
			if i+1 < len(args) {
				config.OutputFile = args[i+1]
				i++
			}
		case "--strict":
			config.StrictMode = true
		case "--verbose", "-v":
			config.Verbose = true
		}
	}

	return performValidation(config)
}

func validateWithRules(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("rules file and data file required")
	}

	// Load validation rules
	rulesData, err := os.ReadFile(args[0])
	if err != nil {
		return fmt.Errorf("failed to read rules file: %w", err)
	}

	var rules []ValidationRule
	if err := json.Unmarshal(rulesData, &rules); err != nil {
		return fmt.Errorf("failed to parse rules file: %w", err)
	}

	config := ValidationConfig{
		Rules:     rules,
		InputFile: args[1],
		Format:    "text",
		MaxErrors: 100,
	}

	return performValidation(config)
}

func generateSchema(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("data file required")
	}

	inputFile := args[0]
	data, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Try to parse as JSON
	var jsonData interface{}
	if err := json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("input file must be valid JSON: %w", err)
	}

	schema := generateSchemaFromData(jsonData)
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	fmt.Println(string(schemaJSON))
	return nil
}

func createRules(args []string) error {
	fmt.Println("Creating validation rules interactively...")
	
	scanner := bufio.NewScanner(os.Stdin)
	var rules []ValidationRule

	for {
		fmt.Print("Field name (or 'done' to finish): ")
		if !scanner.Scan() {
			break
		}
		fieldName := strings.TrimSpace(scanner.Text())
		if fieldName == "done" || fieldName == "" {
			break
		}

		rule := ValidationRule{Field: fieldName}

		fmt.Print("Field type (string, number, boolean, email, url, regex): ")
		if scanner.Scan() {
			rule.Type = strings.TrimSpace(scanner.Text())
		}

		fmt.Print("Required? (y/n): ")
		if scanner.Scan() {
			rule.Required = strings.ToLower(strings.TrimSpace(scanner.Text())) == "y"
		}

		if rule.Type == "string" {
			fmt.Print("Min length (optional): ")
			if scanner.Scan() {
				if minLen, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					rule.MinLength = minLen
				}
			}

			fmt.Print("Max length (optional): ")
			if scanner.Scan() {
				if maxLen, err := strconv.Atoi(strings.TrimSpace(scanner.Text())); err == nil {
					rule.MaxLength = maxLen
				}
			}
		}

		if rule.Type == "regex" {
			fmt.Print("Pattern: ")
			if scanner.Scan() {
				rule.Pattern = strings.TrimSpace(scanner.Text())
			}
		}

		fmt.Print("Custom error message (optional): ")
		if scanner.Scan() {
			message := strings.TrimSpace(scanner.Text())
			if message != "" {
				rule.Message = message
			}
		}

		rules = append(rules, rule)
	}

	rulesJSON, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	filename := "validation-rules.json"
	if err := os.WriteFile(filename, rulesJSON, 0644); err != nil {
		return fmt.Errorf("failed to write rules file: %w", err)
	}

	fmt.Printf("Validation rules saved to %s\n", filename)
	return nil
}

func performValidation(config ValidationConfig) error {
	startTime := time.Now()
	
	result := ValidationResult{
		Valid: true,
		Statistics: ValidationStats{},
	}

	// Read input data
	var data []byte
	var err error
	
	if config.InputFile == "-" || config.InputFile == "" {
		data, err = readStdin()
	} else {
		data, err = os.ReadFile(config.InputFile)
	}
	
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	// Validate data
	if config.SchemaFile != "" {
		err = validateAgainstSchema(data, config, &result)
	} else if len(config.Rules) > 0 {
		err = validateAgainstRules(data, config, &result)
	} else {
		err = validateJSONFormat(data, config, &result)
	}

	if err != nil {
		return err
	}

	// Calculate statistics
	result.Statistics.ProcessingTime = time.Since(startTime).String()
	result.Statistics.ErrorCount = len(result.Errors)
	result.Statistics.WarningCount = len(result.Warnings)
	result.Valid = len(result.Errors) == 0

	// Output results
	return outputResults(result, config)
}

func validateJSONFormat(data []byte, config ValidationConfig, result *ValidationResult) error {
	lines := strings.Split(string(data), "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result.Statistics.TotalRecords++
		
		var jsonData interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Path:       fmt.Sprintf("line %d", lineNum+1),
				Message:    fmt.Sprintf("Invalid JSON: %v", err),
				Rule:       "json_format",
				Severity:   "error",
				LineNumber: lineNum + 1,
			})
			result.Statistics.InvalidRecords++
		} else {
			result.Statistics.ValidRecords++
		}

		if len(result.Errors) >= config.MaxErrors {
			break
		}
	}

	return nil
}

func validateAgainstSchema(data []byte, config ValidationConfig, result *ValidationResult) error {
	// Load schema
	schemaData, err := os.ReadFile(config.SchemaFile)
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	var schema Schema
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return fmt.Errorf("failed to parse schema: %w", err)
	}

	// Validate each JSON object
	lines := strings.Split(string(data), "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result.Statistics.TotalRecords++

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Path:       fmt.Sprintf("line %d", lineNum+1),
				Message:    fmt.Sprintf("Invalid JSON: %v", err),
				Rule:       "json_format",
				Severity:   "error",
				LineNumber: lineNum + 1,
			})
			result.Statistics.InvalidRecords++
			continue
		}

		// Validate against schema
		errors := validateObjectAgainstSchema(jsonData, schema, fmt.Sprintf("line %d", lineNum+1))
		if len(errors) > 0 {
			result.Errors = append(result.Errors, errors...)
			result.Statistics.InvalidRecords++
		} else {
			result.Statistics.ValidRecords++
		}

		if len(result.Errors) >= config.MaxErrors {
			break
		}
	}

	return nil
}

func validateAgainstRules(data []byte, config ValidationConfig, result *ValidationResult) error {
	lines := strings.Split(string(data), "\n")
	
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		result.Statistics.TotalRecords++

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(line), &jsonData); err != nil {
			result.Errors = append(result.Errors, ValidationError{
				Path:       fmt.Sprintf("line %d", lineNum+1),
				Message:    fmt.Sprintf("Invalid JSON: %v", err),
				Rule:       "json_format",
				Severity:   "error",
				LineNumber: lineNum + 1,
			})
			result.Statistics.InvalidRecords++
			continue
		}

		// Validate against rules
		errors := validateObjectAgainstRules(jsonData, config.Rules, fmt.Sprintf("line %d", lineNum+1))
		if len(errors) > 0 {
			result.Errors = append(result.Errors, errors...)
			result.Statistics.InvalidRecords++
		} else {
			result.Statistics.ValidRecords++
		}

		if len(result.Errors) >= config.MaxErrors {
			break
		}
	}

	return nil
}

func validateObjectAgainstSchema(obj map[string]interface{}, schema Schema, path string) []ValidationError {
	var errors []ValidationError

	// Check required fields
	for _, required := range schema.Required {
		if _, exists := obj[required]; !exists {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    required,
				Message:  fmt.Sprintf("Required field '%s' is missing", required),
				Rule:     "required",
				Severity: "error",
			})
		}
	}

	// Validate each field
	for fieldName, value := range obj {
		if schemaField, exists := schema.Properties[fieldName]; exists {
			fieldErrors := validateFieldAgainstSchema(fieldName, value, schemaField, path)
			errors = append(errors, fieldErrors...)
		}
	}

	return errors
}

func validateFieldAgainstSchema(fieldName string, value interface{}, field SchemaField, path string) []ValidationError {
	var errors []ValidationError

	// Type validation
	if !isValidType(value, field.Type) {
		errors = append(errors, ValidationError{
			Path:     path,
			Field:    fieldName,
			Value:    fmt.Sprintf("%v", value),
			Message:  fmt.Sprintf("Expected type '%s', got '%T'", field.Type, value),
			Rule:     "type",
			Severity: "error",
		})
		return errors
	}

	// String validations
	if field.Type == "string" {
		str := fmt.Sprintf("%v", value)
		
		if field.MinLength != nil && len(str) < *field.MinLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    fieldName,
				Value:    str,
				Message:  fmt.Sprintf("String length %d is less than minimum %d", len(str), *field.MinLength),
				Rule:     "minLength",
				Severity: "error",
			})
		}
		
		if field.MaxLength != nil && len(str) > *field.MaxLength {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    fieldName,
				Value:    str,
				Message:  fmt.Sprintf("String length %d exceeds maximum %d", len(str), *field.MaxLength),
				Rule:     "maxLength",
				Severity: "error",
			})
		}
		
		if field.Pattern != "" {
			if matched, _ := regexp.MatchString(field.Pattern, str); !matched {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    fieldName,
					Value:    str,
					Message:  fmt.Sprintf("String does not match pattern '%s'", field.Pattern),
					Rule:     "pattern",
					Severity: "error",
				})
			}
		}
	}

	// Number validations
	if field.Type == "number" || field.Type == "integer" {
		if num, ok := value.(float64); ok {
			if field.Minimum != nil && num < *field.Minimum {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    fieldName,
					Value:    fmt.Sprintf("%v", value),
					Message:  fmt.Sprintf("Value %g is less than minimum %g", num, *field.Minimum),
					Rule:     "minimum",
					Severity: "error",
				})
			}
			
			if field.Maximum != nil && num > *field.Maximum {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    fieldName,
					Value:    fmt.Sprintf("%v", value),
					Message:  fmt.Sprintf("Value %g exceeds maximum %g", num, *field.Maximum),
					Rule:     "maximum",
					Severity: "error",
				})
			}
		}
	}

	// Enum validation
	if len(field.Enum) > 0 {
		str := fmt.Sprintf("%v", value)
		valid := false
		for _, enumValue := range field.Enum {
			if str == enumValue {
				valid = true
				break
			}
		}
		if !valid {
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    fieldName,
				Value:    str,
				Message:  fmt.Sprintf("Value '%s' is not in allowed values: %v", str, field.Enum),
				Rule:     "enum",
				Severity: "error",
			})
		}
	}

	return errors
}

func validateObjectAgainstRules(obj map[string]interface{}, rules []ValidationRule, path string) []ValidationError {
	var errors []ValidationError

	for _, rule := range rules {
		value, exists := obj[rule.Field]
		
		// Required field check
		if rule.Required && !exists {
			message := rule.Message
			if message == "" {
				message = fmt.Sprintf("Required field '%s' is missing", rule.Field)
			}
			errors = append(errors, ValidationError{
				Path:     path,
				Field:    rule.Field,
				Message:  message,
				Rule:     "required",
				Severity: "error",
			})
			continue
		}

		if !exists {
			continue
		}

		// Type validation
		switch rule.Type {
		case "email":
			if !isValidEmail(fmt.Sprintf("%v", value)) {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    rule.Field,
					Value:    fmt.Sprintf("%v", value),
					Message:  getErrorMessage(rule, "Invalid email format"),
					Rule:     "email",
					Severity: "error",
				})
			}
		case "url":
			if !isValidURL(fmt.Sprintf("%v", value)) {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    rule.Field,
					Value:    fmt.Sprintf("%v", value),
					Message:  getErrorMessage(rule, "Invalid URL format"),
					Rule:     "url",
					Severity: "error",
				})
			}
		case "regex":
			if rule.Pattern != "" {
				if matched, _ := regexp.MatchString(rule.Pattern, fmt.Sprintf("%v", value)); !matched {
					errors = append(errors, ValidationError{
						Path:     path,
						Field:    rule.Field,
						Value:    fmt.Sprintf("%v", value),
						Message:  getErrorMessage(rule, fmt.Sprintf("Does not match pattern '%s'", rule.Pattern)),
						Rule:     "regex",
						Severity: "error",
					})
				}
			}
		}

		// Length validation for strings
		if rule.Type == "string" {
			str := fmt.Sprintf("%v", value)
			if rule.MinLength > 0 && len(str) < rule.MinLength {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    rule.Field,
					Value:    str,
					Message:  getErrorMessage(rule, fmt.Sprintf("Length %d is less than minimum %d", len(str), rule.MinLength)),
					Rule:     "minLength",
					Severity: "error",
				})
			}
			if rule.MaxLength > 0 && len(str) > rule.MaxLength {
				errors = append(errors, ValidationError{
					Path:     path,
					Field:    rule.Field,
					Value:    str,
					Message:  getErrorMessage(rule, fmt.Sprintf("Length %d exceeds maximum %d", len(str), rule.MaxLength)),
					Rule:     "maxLength",
					Severity: "error",
				})
			}
		}
	}

	return errors
}

func generateSchemaFromData(data interface{}) Schema {
	schema := Schema{
		Type:       "object",
		Properties: make(map[string]SchemaField),
	}

	if obj, ok := data.(map[string]interface{}); ok {
		for key, value := range obj {
			schema.Properties[key] = generateSchemaField(value)
		}
	}

	return schema
}

func generateSchemaField(value interface{}) SchemaField {
	field := SchemaField{}

	switch v := value.(type) {
	case string:
		field.Type = "string"
		if isValidEmail(v) {
			field.Format = "email"
		} else if isValidURL(v) {
			field.Format = "uri"
		}
	case float64:
		if v == float64(int64(v)) {
			field.Type = "integer"
		} else {
			field.Type = "number"
		}
	case bool:
		field.Type = "boolean"
	case []interface{}:
		field.Type = "array"
	case map[string]interface{}:
		field.Type = "object"
	default:
		field.Type = "string"
	}

	return field
}

func isValidType(value interface{}, expectedType string) bool {
	switch expectedType {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := value.(float64)
		return ok
	case "integer":
		if num, ok := value.(float64); ok {
			return num == float64(int64(num))
		}
		return false
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	default:
		return true
	}
}

func isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched
}

func isValidURL(url string) bool {
	pattern := `^https?://[^\s/$.?#].[^\s]*$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

func getErrorMessage(rule ValidationRule, defaultMessage string) string {
	if rule.Message != "" {
		return rule.Message
	}
	return defaultMessage
}

func readStdin() ([]byte, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return []byte(strings.Join(lines, "\n")), scanner.Err()
}

func outputResults(result ValidationResult, config ValidationConfig) error {
	switch config.Format {
	case "json":
		return outputJSON(result, config)
	case "csv":
		return outputCSV(result, config)
	default:
		return outputText(result, config)
	}
}

func outputJSON(result ValidationResult, config ValidationConfig) error {
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	if config.OutputFile != "" {
		return os.WriteFile(config.OutputFile, jsonData, 0644)
	}

	fmt.Println(string(jsonData))
	return nil
}

func outputCSV(result ValidationResult, config ValidationConfig) error {
	var lines []string
	lines = append(lines, "Path,Field,Value,Message,Rule,Severity,LineNumber")

	for _, err := range result.Errors {
		line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d",
			err.Path, err.Field, err.Value, err.Message, err.Rule, err.Severity, err.LineNumber)
		lines = append(lines, line)
	}

	if config.ShowWarnings {
		for _, warn := range result.Warnings {
			line := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%d",
				warn.Path, warn.Field, warn.Value, warn.Message, warn.Rule, warn.Severity, warn.LineNumber)
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")

	if config.OutputFile != "" {
		return os.WriteFile(config.OutputFile, []byte(content), 0644)
	}

	fmt.Println(content)
	return nil
}

func outputText(result ValidationResult, config ValidationConfig) error {
	var output strings.Builder

	// Summary
	output.WriteString(fmt.Sprintf("Validation Summary:\n"))
	output.WriteString(fmt.Sprintf("  Total Records: %d\n", result.Statistics.TotalRecords))
	output.WriteString(fmt.Sprintf("  Valid Records: %d\n", result.Statistics.ValidRecords))
	output.WriteString(fmt.Sprintf("  Invalid Records: %d\n", result.Statistics.InvalidRecords))
	output.WriteString(fmt.Sprintf("  Errors: %d\n", result.Statistics.ErrorCount))
	output.WriteString(fmt.Sprintf("  Warnings: %d\n", result.Statistics.WarningCount))
	output.WriteString(fmt.Sprintf("  Processing Time: %s\n", result.Statistics.ProcessingTime))
	output.WriteString(fmt.Sprintf("  Overall Status: %s\n\n", getStatusText(result.Valid)))

	// Errors
	if len(result.Errors) > 0 {
		output.WriteString("Errors:\n")
		for _, err := range result.Errors {
			output.WriteString(fmt.Sprintf("  [ERROR] %s", err.Path))
			if err.Field != "" {
				output.WriteString(fmt.Sprintf(".%s", err.Field))
			}
			if err.Value != "" {
				output.WriteString(fmt.Sprintf(" (value: %s)", err.Value))
			}
			output.WriteString(fmt.Sprintf(": %s (%s)\n", err.Message, err.Rule))
		}
		output.WriteString("\n")
	}

	// Warnings
	if config.ShowWarnings && len(result.Warnings) > 0 {
		output.WriteString("Warnings:\n")
		for _, warn := range result.Warnings {
			output.WriteString(fmt.Sprintf("  [WARN] %s", warn.Path))
			if warn.Field != "" {
				output.WriteString(fmt.Sprintf(".%s", warn.Field))
			}
			if warn.Value != "" {
				output.WriteString(fmt.Sprintf(" (value: %s)", warn.Value))
			}
			output.WriteString(fmt.Sprintf(": %s (%s)\n", warn.Message, warn.Rule))
		}
		output.WriteString("\n")
	}

	content := output.String()

	if config.OutputFile != "" {
		return os.WriteFile(config.OutputFile, []byte(content), 0644)
	}

	fmt.Print(content)
	return nil
}

func getStatusText(valid bool) string {
	if valid {
		return "VALID"
	}
	return "INVALID"
}