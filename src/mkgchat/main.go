package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/google/uuid"
)

func main() {

	// Get command line arguments
	args := os.Args[1:]
	var text, title, thread string

	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--text" && i+1 < len(args) {
			text = args[i+1]
		} else if args[i] == "--title" && i+1 < len(args) {
			title = args[i+1]
		} else if args[i] == "--thread" && i+1 < len(args) {
			title = args[i+1]
		}
	}

	if text == "" {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			// Get the line from stdin
			line := scanner.Text()
			create(title, line, thread)

		}
		return
	}

	create(title, text, thread)

}
func create(title string, text string, thread string) {
	if thread == "" {
		thread = uuid.New().String()
		thread = thread[0:8] + thread[9:13] + thread[14:18] + thread[19:23] + thread[24:]
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
		fmt.Fprintln(os.Stderr, "Error marshaling payload:", err)
		return
	}
	// Write the payload JSON to stdout
	fmt.Fprintln(os.Stdout, string(payloadJSON))

}
