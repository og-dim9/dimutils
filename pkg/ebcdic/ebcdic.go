package ebcdic

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Config holds configuration for EBCDIC operations
type Config struct {
	Codepage string
	Encode   bool
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		Codepage: "037",
		Encode:   false,
	}
}

// Run processes stdin with EBCDIC encoding/decoding
func Run(args []string) error {
	config := DefaultConfig()
	
	// Parse arguments
	for i, arg := range args {
		switch arg {
		case "-c", "--codepage":
			if i+1 < len(args) {
				config.Codepage = args[i+1]
			}
		case "-e", "--encode":
			config.Encode = true
		case "-h", "--help":
			printHelp()
			return nil
		}
	}

	codePage, err := getCodePage(config.Codepage)
	if err != nil {
		return fmt.Errorf("error parsing codepage: %w", err)
	}

	return processInput(codePage, config.Encode)
}

func printHelp() {
	fmt.Println("Usage: ebcdic [options]")
	fmt.Println("Options:")
	fmt.Println("  -c, --codepage CODEPAGE  EBCDIC-Codepage to use (supported: 037 / 273 / 500 / 1140 / 1141 / 1148)")
	fmt.Println("  -e, --encode             Encode input instead of decoding it")
	fmt.Println("  -h, --help               Show this help message")
}

func processInput(codePage int, encode bool) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if encode {
			fmt.Fprintf(os.Stderr, "EBCDIC encoding not yet implemented\n")
			continue
		} else {
			fmt.Fprintf(os.Stderr, "EBCDIC decoding not yet implemented\n") 
			continue
		}
	}
	
	return scanner.Err()
}

func getCodePage(codepage string) (int, error) {
	switch strings.ToLower(codepage) {
	case "37", "037", "ebcdic037":
		return 37, nil
	case "273", "ebcdic273":
		return 273, nil
	case "500", "ebcdic500":
		return 500, nil
	case "1140", "ebcdic1140":
		return 1140, nil
	case "1141", "ebcdic1141":
		return 1141, nil
	case "1148", "ebcdic1148":
		return 1148, nil
	default:
		return 0, fmt.Errorf("unsupported codepage %s", codepage)
	}
}