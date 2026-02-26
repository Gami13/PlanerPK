package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	inPath := "testdata/classTextsParts.txt"
	fourPath := "testdata/classTextsParts4.txt"
	restPath := "testdata/classTextsPartsRest.txt"

	file, err := os.Open(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	var fourParts, rest []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		var parts []string
		if err := json.Unmarshal([]byte(line), &parts); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing line: %v\n", err)
			continue
		}
		if len(parts) == 4 {
			fourParts = append(fourParts, line)
		} else {
			rest = append(rest, line)
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	writeLines(fourPath, fourParts)
	writeLines(restPath, rest)

	fmt.Fprintf(os.Stderr, "4 parts: %d -> %s\n", len(fourParts), fourPath)
	fmt.Fprintf(os.Stderr, "Rest: %d -> %s\n", len(rest), restPath)
}

func writeLines(path string, lines []string) {
	out, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", path, err)
		os.Exit(1)
	}
	defer out.Close()
	for _, line := range lines {
		fmt.Fprintln(out, line)
	}
}
