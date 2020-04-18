package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type inputFile struct {
	filepath  string
	separator string
	pretty    bool
}

func getFileData() (*inputFile, error) {
	// Validate arguments
	if len(os.Args) < 2 {
		return nil, errors.New("A filepath argument is required")
	}

	separator := flag.String("separator", "comma", "Column separator (Default is comma)")
	pretty := flag.Bool("pretty", false, "Generate pretty JSON (Default is false)")

	flag.Parse()

	fileLocation := flag.Arg(0)

	// Check if file does not exist
	if _, err := os.Stat(fileLocation); err != nil && os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("File %s does not exist", fileLocation))
	}

	// Check if file is CSV
	if fileExtension := filepath.Ext(fileLocation); fileExtension != ".csv" {
		return nil, errors.New(fmt.Sprintf("File %s is not CSV", fileLocation))
	}

	if !(*separator == "comma" || *separator == "semicolon") {
		return nil, errors.New("Only comma or semicolon separators are allowed")
	}

	return &inputFile{fileLocation, *separator, *pretty}, nil
}

func processLine(headers []string, rawLine string, separator string) (map[string]string, error) {
	dataList := strings.Split(rawLine, separator)

	if len(dataList) != len(headers) {
		return nil, errors.New("Line doesn't match headers format. Skipping.")
	}

	recordMap := make(map[string]string)

	for i, name := range headers {
		recordMap[name] = dataList[i]
	}

	return recordMap, nil
}

func processFile(fileData *inputFile, writerChannel <-chan map[string]string) {
	file, err := os.Open(fileData.filepath)

	if err != nil {
		panic(err)
	}

	defer file.Close()

	separator := ","
	if fileData.separator == "semicolon" {
		separator = ";"
	}

	// Get Headers
	reader := bufio.NewReader(file)
	var line string
	line, err = reader.ReadString('\n')

	if err != nil {
		panic("Invalid csv content")
	}

	headers := strings.Split(line, separator)
	fmt.Println(headers)

	for {
		line, err = reader.ReadString('\n')

		if err != nil {
			break
		}

		record, err := processLine(headers, line, separator)

		if err != nil {
			fmt.Printf("Line: %sError: %s\n", line, err)
			continue
		}
		// Process the line here.
		fmt.Println(record)
	}
}

func main() {
	fileData, err := getFileData()

	if err != nil {
		panic(err)
	}

	writerChannel := make(chan map[string]string)

	processFile(fileData, writerChannel)
}
