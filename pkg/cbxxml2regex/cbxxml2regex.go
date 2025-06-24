package cbxxml2regex

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Condition represents XML condition element
type Condition struct {
	Name    string      `xml:"name,attr"`
	Through string      `xml:"through,attr"`
	Value   string      `xml:"value,attr"`
	Items   []Condition `xml:"condition"`
}

// Copybook represents XML copybook element
type Copybook struct {
	Filename string `xml:"filename,attr"`
	Items    []Item `xml:"item"`
}

// Item represents XML item element
type Item struct {
	AssumedDigits      int         `xml:"assumed-digits,attr"`
	DependingOn        string      `xml:"depending-on,attr"`
	DisplayLength      int         `xml:"display-length,attr"`
	EdittedNumeric     bool        `xml:"editted-numeric,attr"`
	InheritedUsage     bool        `xml:"inherited-usage,attr"`
	InsertDecimalPoint bool        `xml:"insert-decimal-point,attr"`
	Level              string      `xml:"level,attr"`
	Name               string      `xml:"name,attr"`
	Numeric            string      `xml:"numeric,attr"`
	Occurs             int         `xml:"occurs,attr"`
	OccursMin          int         `xml:"occurs-min,attr"`
	Picture            string      `xml:"picture,attr"`
	Position           int         `xml:"position,attr"`
	Redefined          string      `xml:"redefined,attr"`
	Redefines          string      `xml:"redefines,attr"`
	Scale              int         `xml:"scale,attr"`
	SignPosition       string      `xml:"sign-position,attr"`
	SignSeparate       bool        `xml:"sign-separate,attr"`
	Signed             bool        `xml:"signed,attr"`
	StorageLength      int         `xml:"storage-length,attr"`
	Sync               bool        `xml:"sync,attr"`
	Usage              string      `xml:"usage,attr"`
	Value              string      `xml:"value,attr"`
	Conditions         []Condition `xml:"condition"`
	Items              []Item      `xml:"item"`
}

// Run processes COBOL XML files to generate regex patterns
func Run(args []string) error {
	var xmlFilename string
	
	// Parse arguments
	for i, arg := range args {
		switch arg {
		case "--xml":
			if i+1 < len(args) {
				xmlFilename = args[i+1]
			}
		case "-h", "--help":
			printHelp()
			return nil
		}
	}

	if xmlFilename == "" {
		return fmt.Errorf("please provide the XML file path using the --xml flag")
	}

	return processXMLFile(xmlFilename)
}

func printHelp() {
	fmt.Println("Usage: cbxxml2regex --xml <file.xml>")
	fmt.Println("Convert COBOL XML definitions to regular expressions")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  --xml FILE    XML file path")
	fmt.Println("  -h, --help    Show this help message")
}

func processXMLFile(xmlFilename string) error {
	xmlFile, err := os.Open(xmlFilename)
	if err != nil {
		return fmt.Errorf("error opening XML file: %w", err)
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		return fmt.Errorf("error reading XML content: %w", err)
	}

	var copybook Copybook
	err = xml.Unmarshal(byteValue, &copybook)
	if err != nil {
		return fmt.Errorf("error parsing XML: %w", err)
	}

	fmt.Println("Copybook Filename:", copybook.Filename)
	items := flatten(copybook.Items)
	regex := "^"
	
	for i, item := range items {
		itemName := formatItemName(item.Name, copybook.Filename)
		if i != len(items)-1 {
			length := items[i+1].Position - item.Position
			if length != 0 {
				regex += fmt.Sprintf("(?P<%s>.{%d})", itemName, length)
			}
		} else {
			regex += fmt.Sprintf("(?P<%s>.*)", itemName)
		}
	}
	
	fmt.Println(regex)
	return nil
}

func formatItemName(name, copybookName string) string {
	name = strings.ReplaceAll(name, copybookName, "")
	name = strings.ReplaceAll(name, "-", " ")
	finalName := ""
	for _, part := range strings.Split(name, " ") {
		finalName += strings.Title(strings.ToLower(part))
	}
	return finalName
}

func flatten(items []Item) []Item {
	var result []Item
	for _, item := range items {
		result = append(result, item)
		if len(item.Items) > 0 {
			result = append(result, flatten(item.Items)...)
		}
	}
	return result
}