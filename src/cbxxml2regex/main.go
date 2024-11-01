package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {

	xmlFilename := flag.String("xml", "", "XML file path")
	flag.Parse()

	if *xmlFilename == "" {
		fmt.Println("Please provide the XML file path using the --xml flag")
		return
	}

	xmlFile, err := os.Open(*xmlFilename)
	if err != nil {
		fmt.Println("Error opening XML file:", err)
		return
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		fmt.Println("Error reading XML content:", err)
		return
	}

	var copybook Copybook
	err = xml.Unmarshal(byteValue, &copybook)
	if err != nil {
		fmt.Println("Error parsing XML:", err)
		return
	}

	fmt.Println("Copybook Filename:", copybook.Filename)
	items := flatten(copybook.Items)
	regex := "^"
	for i, item := range items {
		in := itemName(item.Name, copybook.Filename)
		if i != len(items)-1 {
			length := items[i+1].Position - item.Position
			if length != 0 {
				// fmt.Println(i, item.Name, length, item.Position)
				regex += fmt.Sprintf("(?P<%s>.{%d})", in, length)
				// regex += fmt.Sprintf("(.{%d})", length)
			}
		} else {
			regex += fmt.Sprintf("(?P<%s>.*)", in)
		}
	}
	fmt.Println(regex)
}
func itemName(name, copybookName string) string {
	name = strings.ReplaceAll(name, copybookName, "")
	name = strings.ReplaceAll(name, "-", " ")
	finalName := ""
	for _, part := range strings.Split(name, " ") {
		finalName += strings.Title(strings.ToLower(part))
	}
	return finalName
}

func getregexpart(items []Item, level int, copybookName string) []string {
	var result []string
	for _, item := range items {
		if len(item.Items) > 0 {
			result = append(result, getregexpart(item.Items, level+1, copybookName)...)
		} else {
			name := strings.ReplaceAll(item.Name, copybookName+"-", "")
			result = append(result, fmt.Sprintf("%s (%d)", name, item.Position))
		}
	}
	return result
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

// func getitems(items []Item, level int) []string {
// 	var result []string
// 	for _, item := range items {
// 		if len(item.Items) > 0 {
// 			result = append(result, fmt.Sprintf("-%s%s %s (%d)", getIndent(level), item.Level, item.Name, item.Position))

// 			result = append(result, getitems(item.Items, level+1)...)
// 		} else {
// 			result = append(result, fmt.Sprintf("%s%s %s (%d)", getIndent(level), item.Level, item.Name, item.Position))
// 		}
// 	}
// 	return result
// }

// func getIndent(level int) string {
// 	var indent string
// 	for i := 0; i < level; i++ {
// 		indent += "  "
// 	}
// 	return indent
// }
