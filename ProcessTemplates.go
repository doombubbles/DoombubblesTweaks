package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v2"
)

type Values map[string]interface{}

func main() {
	// Read the values from values.yaml
	var values Values = nil
	valuesData, err := os.ReadFile("values.yaml")
	if err == nil {
		err = yaml.Unmarshal(valuesData, &values)
		if err != nil {
			log.Fatalf("error parsing values.yaml: %v", err)
		}
	}

	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		return processFile(path, info, err, values)
	})
	if err != nil {
		log.Fatalf("error walking the path: %v", err)
	}
}

func processFile(path string, info os.FileInfo, err error, values Values) error {
	if err != nil {
		log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		return err
	}
	if info.IsDir() {
		return nil // Skip directories
	}
	if filepath.Ext(path) != ".tmpl" {
		return nil // Skip non-tmpl files
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("error reading file %s: %v", path, err)
		return err
	}

	var extension string
	if strings.Contains(path, ".loca") {
		extension = ".xml"
	} else if strings.Contains(path, ".lsf") {
		extension = ".lsx"
	} else {
		extension = ".txt"
	}

	var text = string(content)
	if extension == ".txt" {
		text = fmt.Sprintf("// Generated by %s using ProcessTemplates.go\n", filepath.Base(path)) + text
	} else {
		text = strings.Replace(text, "<?xml version=\"1.0\" encoding=\"utf-8\"?>",
			fmt.Sprintf("<?xml version=\"1.0\" encoding=\"utf-8\"?> <!-- Generated by %s using ProcessTemplates.go -->", filepath.Base(path)), 1)
	}

	// Parse and execute template
	tmpl, err := template.New("template").Funcs(sprig.FuncMap()).Parse(text)
	if err != nil {
		log.Printf("error parsing template %s: %v", path, err)
		return err
	}

	outputPath := strings.TrimSuffix(path, ".tmpl") + extension
	outputFile, err := os.Create(outputPath)
	if err != nil {
		log.Printf("error creating output file %s: %v", outputPath, err)
		return err
	}
	defer outputFile.Close()

	// Write an additional line to the top of the file
	if err != nil {
		log.Printf("error writing to output file %s: %v", outputPath, err)
		return err
	}

	// Execute the template and write to the output file
	err = tmpl.Execute(outputFile, values) // You can pass any data you want to template here
	if err != nil {
		log.Printf("error executing template for file %s: %v", path, err)
		return err
	}

	log.Printf("Processed and created: %s", outputPath)
	return nil
}
