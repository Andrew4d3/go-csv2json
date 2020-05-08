# go-csv2json

This is a tool to transform a CSV file into a JSON one. It will parse all CSV rows from a file into a JSON array and output the result into a `.json` file. Each data cell will be parsed as a string datatype.

## How to use it
```
$ ./csv2json [options] <csvFilePath>
```
Example:
```
$ ./csv2json data.csv
```
The command above will create a `data.json` file at exactly the same dir location of `data.csv`
## Options Available
- **pretty**: if enabled, it will create a well-formatted JSON file instead of a compact one (one single line)
- **separator**: to indicate which character is used to separate cells. Only accepted options are `comma` (default) or `semicolon`.

Example using options:
```
$ ./csv2json --pretty --separator=semicolon data.csv
```
The command above will create a formatted `data.json` file at exactly the same dir location of `data.csv` The row columns from this file are separated using semicolons instead of commas.

## Learning Go
I made this tool with the only intention to practice my recently acquired Golang skills. Feel free to use the source code as a way of learning more about how to use this programming language. Especially for creating CLI tools.

## License

MIT