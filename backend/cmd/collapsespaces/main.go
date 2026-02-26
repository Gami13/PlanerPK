package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

var manySpaces = regexp.MustCompile(` {5,}`)

func main() {
	path := "testdata/classTexts.txt"

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, manySpaces.ReplaceAllString(scanner.Text(), "     "))
	}
	file.Close()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}

	out, err := os.Create(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing file: %v\n", err)
		os.Exit(1)
	}
	defer out.Close()

	for _, line := range lines {
		fmt.Fprintln(out, line)
	}

	fmt.Fprintf(os.Stderr, "Collapsed spaces in %d lines\n", len(lines))
}
