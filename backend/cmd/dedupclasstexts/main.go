package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	path := "testdata/classTexts.txt"

	file, err := os.Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}

	seen := make(map[string]struct{})
	var unique []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if _, ok := seen[line]; !ok {
			seen[line] = struct{}{}
			unique = append(unique, line)
		}
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

	for _, line := range unique {
		fmt.Fprintln(out, line)
	}

	fmt.Fprintf(os.Stderr, "Deduplicated: %d -> %d entries\n", len(unique)+len(seen)-len(unique), len(unique))
}
