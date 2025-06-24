package regex2json

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Config holds configuration for regex2json
type Config struct {
	RemoveEmpty bool
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		RemoveEmpty: true,
	}
}

// Run processes stdin with the given regex pattern
func Run(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no regex pattern provided. Usage: regex2json 'regex pattern'")
	}
	
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Println("Usage: cat file.txt | regex2json 'regex pattern'")
		return nil
	}
	
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	pattern, err := regexp.Compile(args[0])
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	config := DefaultConfig()
	return processInput(pattern, config)
}

func processInput(pattern *regexp.Regexp, config Config) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		result := make(map[string]string)
		matches := pattern.FindAllStringSubmatch(scanner.Text(), -1)
		
		for _, match := range matches {
			for i, name := range pattern.SubexpNames() {
				if name == "" {
					name = fmt.Sprintf("match_%d", i)
				}
				if i != 0 {
					result[name] = strings.TrimSpace(match[i])

					if config.RemoveEmpty && result[name] == "" {
						delete(result, name)
					}
				}
			}
		}
		
		if err := printJSON(result); err != nil {
			return err
		}
	}
	
	return scanner.Err()
}

func printJSON(result map[string]string) error {
	jsonString, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}
	fmt.Println(string(jsonString))
	return nil
}