package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

// logger flags
var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime|log.Lshortfile)

func usage() {
	fmt.Println("Usage: togchat --spaceid <spaceid> --key <apikey> --token <token>")
	fmt.Println("  or:  togchat --help -h for usage")
	fmt.Println("  ENVIRONMENT VARIABLES: GCHAT_SPACEID, GCHAT_KEY, GCHAT_TOKEN")
	os.Exit(1)
}
func main() {

	// Define your Google Chat webhook URL

	args := os.Args[1:]
	var spaceid, apikey, token string

	// Parse command line arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "--spaceid" && i+1 < len(args) {
			spaceid = args[i+1]
		} else if args[i] == "--key" && i+1 < len(args) {
			apikey = args[i+1]
		} else if args[i] == "--token" && i+1 < len(args) {
			apikey = args[i+1]
		} else if args[i] == "--help" && i+1 < len(args) {
			usage()
		} else if args[i] == "-h" && i+1 < len(args) {
			usage()
		}
	}

	// fall back to environment variables
	if spaceid == "" {
		spaceid = os.Getenv("GCHAT_SPACEID")
	}
	if apikey == "" {
		apikey = os.Getenv("GCHAT_KEY")
	}
	if token == "" {
		token = os.Getenv("GCHAT_TOKEN")
	}
	// print uage if no arguments are provided
	if spaceid == "" || apikey == "" || token == "" {
		usage()
	}

	done := make(chan bool)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		go sendLine(scanner.Bytes(), spaceid, apikey, token, done)
	}
	<-done
}

func sendLine(body []byte, spaceid string, apikey string, token string, done chan bool) {

	//check if body is a valid json
	if body[0] != '{' {
		msg := map[string]interface{}{"text": string(body), "thread": map[string]string{"threadKey": uuid.New().String()}}
		body, _ = json.Marshal(msg)
	}

	logger.Println("Converting to JSON:", string(body))
	// Send payload to webhook
	webhookURL := "https://chat.googleapis.com/v1/spaces/" + spaceid + "/messages?messageReplyOption=REPLY_MESSAGE_FALLBACK_TO_NEW_THREAD&key=" + apikey + "&token=" + token

	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		logger.Println("Failed to send payload to webhook:", err)
		done <- true
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Println("Failed to send payload to webhook:", resp.Status)
		done <- true
		return
	}

	logger.Println("Payload sent successfully!:", resp.Status)

	done <- true

}
