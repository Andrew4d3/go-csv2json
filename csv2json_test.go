package main

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func Test_getFileData(t *testing.T) {
	tests := []struct {
		name    string
		want    *inputFile
		wantErr bool
		osArgs  []string
	}{
		{"Default parameters", &inputFile{"test.csv", "comma", false}, false, []string{"cmd", "test.csv"}},
		{"No parameters", nil, true, []string{"cmd"}},
		{"Semicolon enabled", &inputFile{"test.csv", "semicolon", false}, false, []string{"cmd", "--separator=semicolon", "test.csv"}},
		{"Pretty enabled", &inputFile{"test.csv", "comma", true}, false, []string{"cmd", "--pretty", "test.csv"}},
		{"Pretty and semicolon enabled", &inputFile{"test.csv", "semicolon", true}, false, []string{"cmd", "--pretty", "--separator=semicolon", "test.csv"}},
		{"Separator not identified", nil, true, []string{"cmd", "--separator=pipe", "test.csv"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualOsArgs := os.Args
			defer func() {
				os.Args = actualOsArgs
				flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) //flags are now reset
			}()

			os.Args = tt.osArgs
			got, err := getFileData()
			if (err != nil) != tt.wantErr {
				t.Errorf("getFileData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFileData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkIfValidFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", "test*.csv")
	if err != nil {
		panic(err)
	}

	defer os.Remove(tmpfile.Name())

	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{"File does exist", args{tmpfile.Name()}, true, false},
		{"File does not exist", args{"nowhere/test.csv"}, false, true},
		{"File is not csv", args{"test.txt"}, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIfValidFile(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("checkIfValidFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("checkIfValidFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_processCsvFile(t *testing.T) {

	wantMapSlice := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}

	tests := []struct {
		name      string
		csvString string
		separator string
	}{
		{"Comma separator", "COL1,COL2,COL3\n1,2,3\n4,5,6\n", "comma"},
		{"Semicolon separator", "COL1;COL2;COL3\n1;2;3\n4;5;6\n", "semicolon"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpfile, err := ioutil.TempFile("", "test*.csv")
			check(err)

			defer os.Remove(tmpfile.Name())
			_, err = tmpfile.WriteString(tt.csvString)
			tmpfile.Sync()

			testFileData := inputFile{
				filepath:  tmpfile.Name(),
				pretty:    false,
				separator: tt.separator,
			}

			writerChannel := make(chan map[string]string)

			go processCsvFile(&testFileData, writerChannel)

			for _, wantMap := range wantMapSlice {
				record := <-writerChannel
				if !reflect.DeepEqual(record, wantMap) {
					t.Errorf("processCsvFile() = %v, want %v", record, wantMap)
				}
			}
		})
	}
}

func Test_writeJSONFile(t *testing.T) {
	dataMap := []map[string]string{
		{"COL1": "1", "COL2": "2", "COL3": "3"},
		{"COL1": "4", "COL2": "5", "COL3": "6"},
	}
	tests := []struct {
		csvPath  string
		jsonPath string
		pretty   bool
		name     string
	}{
		{"compact.csv", "compact.json", false, "Compact JSON"},
		{"pretty.csv", "pretty.json", true, "Pretty JSON"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			writerChannel := make(chan map[string]string)
			done := make(chan bool)

			go func() {
				for _, record := range dataMap {
					writerChannel <- record
				}
				close(writerChannel)
			}()

			go writeJSONFile(tt.csvPath, writerChannel, done, tt.pretty)

			<-done

			testOutput, err := ioutil.ReadFile(tt.jsonPath)

			if err != nil {
				t.Errorf("writeJSONFile(), Output file got error: %v", err)
			}

			defer os.Remove(tt.jsonPath)

			wantOutput, err := ioutil.ReadFile(filepath.Join("testJsonFiles", tt.jsonPath))
			check(err)

			if (string(testOutput)) != (string(wantOutput)) {
				t.Errorf("writeJSONFile() = %v, want %v", string(testOutput), string(wantOutput))
			}
		})
	}
}
