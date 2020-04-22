package main

import (
	"bufio"
	"encoding/json"
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

func check(e error) {
	if e != nil {
		panic(e)
	}
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

func processCsvFile(fileData *inputFile, writerChannel chan<- map[string]string) {
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

	for {
		line, err = reader.ReadString('\n')

		if err != nil {
			close(writerChannel)
			break
		}

		record, err := processLine(headers, line, separator)

		if err != nil {
			fmt.Printf("Line: %sError: %s\n", line, err)
			continue
		}

		writerChannel <- record
	}
}

func createStringWriter(csvPath string) func(string, bool) {
	jsonDir := filepath.Dir(csvPath)
	jsonName := fmt.Sprintf("%s.json", strings.TrimSuffix(filepath.Base(csvPath), ".csv"))
	finalLocation := fmt.Sprintf("%s/%s", jsonDir, jsonName)

	f, err := os.Create(finalLocation)
	check(err)

	return func(data string, close bool) {
		_, err := f.WriteString(data)
		check(err)

		if close {
			f.Close()
		}
	}
}

func writeJsonFile(csvPath string, writerChannel <-chan map[string]string, done chan<- bool) {
	writeString := createStringWriter(csvPath)
	fmt.Println("Writing JSON file...")
	writeString("[", false)
	first := true
	for {
		record, more := <-writerChannel
		if more {
			if !first {
				writeString(",", false)
			} else {
				first = false
			}

			jsonData, _ := json.Marshal(record)
			writeString(string(jsonData), false)
		} else {
			writeString("]", true)
			fmt.Println("Completed!")
			done <- true
			break
		}
	}
}

func main() {
	fileData, err := getFileData()

	if err != nil {
		panic(err)
	}

	writerChannel := make(chan map[string]string)
	done := make(chan bool)

	go processCsvFile(fileData, writerChannel)
	go writeJsonFile(fileData.filepath, writerChannel, done)

	<-done
}
