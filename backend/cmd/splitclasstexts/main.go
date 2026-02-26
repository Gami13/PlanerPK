package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var splitRegex = regexp.MustCompile(`( {4,}|\\n)+`)

func main() {
	inPath := "testdata/classTexts.txt"
	outPath := "testdata/classTextsParts.txt"

	file, err := os.Open(inPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	var lines [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		// Strip surrounding quotes from %q format
		line = strings.TrimPrefix(line, "\"")
		line = strings.TrimSuffix(line, "\"")

		parts := splitRegex.Split(line, -1)

		var cleaned []string
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				cleaned = append(cleaned, p)
			}
		}

		if len(cleaned) > 0 {
			lines = append(lines, cleaned)
		}
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	out, err := os.Create(outPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)
	for _, parts := range lines {
		if err := enc.Encode(parts); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Fprintf(os.Stderr, "Split %d entries -> %s\n", len(lines), outPath)
}
