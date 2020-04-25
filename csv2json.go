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
		return nil, fmt.Errorf("File %s does not exist", fileLocation)
	}

	return &inputFile{fileLocation, *separator, *pretty}, nil
}

func checkIfValidFile(filename string) (bool, error) {
	// Check if file does exist
	if _, err := os.Stat(filename); err != nil && os.IsNotExist(err) {
		return false, fmt.Errorf("File %s does not exist", filename)
	}

	// Check if file is CSV
	if fileExtension := filepath.Ext(filename); fileExtension != ".csv" {
		return false, fmt.Errorf("File %s is not CSV", filename)
	}

	return true, nil
}

func processLine(headers []string, rawLine string, separator string) (map[string]string, error) {
	dataList := strings.Split(rawLine, separator)

	if len(dataList) != len(headers) {
		return nil, errors.New("Line doesn't match headers format. Skipping")
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

func writeJSONFile(csvPath string, writerChannel <-chan map[string]string, done chan<- bool, pretty bool) {
	writeString := createStringWriter(csvPath)

	var jsonFunc func(map[string]string) string
	var breakLine string
	if pretty {
		breakLine = "\n"
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.MarshalIndent(record, "   ", "   ")
			return "   " + string(jsonData)
		}
	} else {
		breakLine = ""
		jsonFunc = func(record map[string]string) string {
			jsonData, _ := json.Marshal(record)
			return string(jsonData)
		}
	}

	fmt.Println("Writing JSON file...")

	writeString("["+breakLine, false)
	first := true
	for {
		record, more := <-writerChannel
		if more {
			if !first {
				writeString(","+breakLine, false)
			} else {
				first = false
			}

			jsonData := jsonFunc(record)
			writeString(jsonData, false)
		} else {
			writeString(breakLine+"]", true)
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

	if _, err := checkIfValidFile(fileData.filepath); err != nil {
		panic(err)
	}

	writerChannel := make(chan map[string]string)
	done := make(chan bool)

	go processCsvFile(fileData, writerChannel)
	go writeJSONFile(fileData.filepath, writerChannel, done, fileData.pretty)

	<-done
}
