package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readLines resolves filename across default directories and returns non-empty, trimmed lines.
func readLines(filename string) ([]string, error) {
	path := filename
	for _, dir := range []string{"data", "../data", "../../data", "."} {
		p := filepath.Join(dir, filename)
		if _, err := os.Stat(p); err == nil {
			path = p
			break
		}
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	return lines, scanner.Err()
}

func main() {
	classrooms, err := readLines("classrooms_legend.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	inputs, err := readLines("all_classrooms.txt")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	for _, input := range inputs {
		result := MatchLine(input, classrooms)
		fmt.Printf("%s -> %s\n", result.Input, result.Output)
	}
}
