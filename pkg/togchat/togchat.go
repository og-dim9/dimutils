package togchat

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/uuid"
)

// Config holds configuration for sending to Google Chat
type Config struct {
	SpaceID string
	APIKey  string
	Token   string
}

// Run sends messages to Google Chat
func Run(args []string) error {
	config := Config{}
	
	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--spaceid":
			if i+1 < len(args) {
				config.SpaceID = args[i+1]
				i++
			}
		case "--key":
			if i+1 < len(args) {
				config.APIKey = args[i+1]
				i++
			}
		case "--token":
			if i+1 < len(args) {
				config.Token = args[i+1]
				i++
			}
		case "-h", "--help":
			printHelp()
			return nil
		}
	}

	// Fall back to environment variables
	if config.SpaceID == "" {
		config.SpaceID = os.Getenv("GCHAT_SPACEID")
	}
	if config.APIKey == "" {
		config.APIKey = os.Getenv("GCHAT_KEY")
	}
	if config.Token == "" {
		config.Token = os.Getenv("GCHAT_TOKEN")
	}

	// Validate required parameters
	if config.SpaceID == "" || config.APIKey == "" || config.Token == "" {
		return fmt.Errorf("missing required parameters. Use --spaceid, --key, --token or set GCHAT_SPACEID, GCHAT_KEY, GCHAT_TOKEN environment variables")
	}

	return processInput(config)
}

func printHelp() {
	fmt.Println("Usage: togchat --spaceid <spaceid> --key <apikey> --token <token>")
	fmt.Println("Send messages to Google Chat")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --spaceid ID    Google Chat space ID")
	fmt.Println("  --key KEY       API key")
	fmt.Println("  --token TOKEN   Token")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  GCHAT_SPACEID   Space ID")
	fmt.Println("  GCHAT_KEY       API key")
	fmt.Println("  GCHAT_TOKEN     Token")
}

func processInput(config Config) error {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)
	
	var wg sync.WaitGroup
	scanner := bufio.NewScanner(os.Stdin)
	
	for scanner.Scan() {
		wg.Add(1)
		go func(line []byte) {
			defer wg.Done()
			if err := sendLine(line, config, logger); err != nil {
				logger.Printf("Error sending message: %v", err)
			}
		}([]byte(scanner.Text()))
	}
	
	wg.Wait()
	return scanner.Err()
}

func sendLine(body []byte, config Config, logger *log.Logger) error {
	// Check if body is valid JSON, if not wrap it
	if len(body) == 0 || body[0] != '{' {
		msg := map[string]interface{}{
			"text":   string(body),
			"thread": map[string]string{"threadKey": uuid.New().String()},
		}
		var err error
		body, err = json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("error marshaling message: %w", err)
		}
	}

	logger.Println("Converting to JSON:", string(body))

	// Send payload to webhook
	webhookURL := fmt.Sprintf("https://chat.googleapis.com/v1/spaces/%s/messages?messageReplyOption=REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD&key=%s&token=%s",
		config.SpaceID, config.APIKey, config.Token)

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to send payload to webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to send payload to webhook: %s", resp.Status)
	}

	logger.Println("Payload sent successfully:", resp.Status)
	return nil
}