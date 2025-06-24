package mkgchat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

// Config holds configuration for Google Chat message creation
type Config struct {
	Text   string
	Title  string
	Thread string
}

// Run creates Google Chat messages from input
func Run(args []string) error {
	config := Config{}
	
	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--text":
			if i+1 < len(args) {
				config.Text = args[i+1]
				i++
			}
		case "--title":
			if i+1 < len(args) {
				config.Title = args[i+1]
				i++
			}
		case "--thread":
			if i+1 < len(args) {
				config.Thread = args[i+1]
				i++
			}
		case "-h", "--help":
			printHelp()
			return nil
		}
	}

	if config.Text == "" {
		return processStdin(config)
	}

	return createMessage(config.Title, config.Text, config.Thread)
}

func printHelp() {
	fmt.Println("Usage: mkgchat [options]")
	fmt.Println("Create Google Chat message JSON from text input")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --text TEXT     Message text")
	fmt.Println("  --title TITLE   Message title")
	fmt.Println("  --thread ID     Thread ID")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("")
	fmt.Println("If no --text is provided, reads from stdin")
}

func processStdin(config Config) error {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if err := createMessage(config.Title, line, config.Thread); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func createMessage(title, text, thread string) error {
	if thread == "" {
		threadUUID := uuid.New().String()
		thread = threadUUID[0:8] + threadUUID[9:13] + threadUUID[14:18] + threadUUID[19:23] + threadUUID[24:]
	}

	// Create the payload
	payload := map[string]interface{}{
		"thread": map[string]string{"threadKey": thread},
		"cards": []map[string]interface{}{
			{
				"sections": []map[string]interface{}{
					{
						"header": title,
						"widgets": []map[string]interface{}{
							{
								"textParagraph": map[string]interface{}{
									"text": text,
								},
							},
						},
					},
				},
			},
		},
	}

	// Convert the payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("error marshaling payload: %w", err)
	}

	// Write the payload JSON to stdout
	fmt.Println(string(payloadJSON))
	return nil
}