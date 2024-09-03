package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"os"
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
	items := getitems(copybook.Items, 0)
	for _, item := range items {
		fmt.Println(item)
	}

}
