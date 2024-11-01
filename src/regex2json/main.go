package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const (
	remove_empty = true
)

func main() {
	//TODO: switch to gnu style flags
	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Println("No regex pattern provided")
		os.Exit(1)
	}
	if args[0] == "-h" || args[0] == "--help" {
		fmt.Println("Usage: cat file.txt | regex2json 'regex pattern'")
		os.Exit(0)
	}
	if len(args) > 1 {
		fmt.Println("Too many arguments")
		os.Exit(1)
	}

	pattern, err := regexp.Compile(args[0])
	if err != nil {
		fmt.Println("Invalid regex pattern")
		os.Exit(1)
	}

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

					if remove_empty && result[name] == "" {
						delete(result, name)
					}

				}
			}
		}
		printJSON(result)
	}
}

func printJSON(result map[string]string) {

	jsonString, err := json.Marshal(result)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error marshaling JSON:", err)
	} else {
		fmt.Fprintln(os.Stdout, string(jsonString))
	}
}
