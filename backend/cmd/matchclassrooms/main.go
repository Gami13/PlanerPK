package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/woodpkrs/PlanerPK/classroommatch"
)

// readLines resolves filename across default directories and returns non-empty, trimmed lines.
func readLines(filename string) ([]string, error) {
	path := filename
	for _, dir := range []string{"testdata", "data", "../data", "../../data", "."} {
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

	candidates := make([]classroommatch.Candidate, 0, len(classrooms))
	for _, classroom := range classrooms {
		candidates = append(candidates, classroommatch.Candidate{
			ClassroomKey: classroom,
			BuildingKey:  classroom,
		})
	}
	matcher := classroommatch.New(candidates, nil)
	for _, input := range inputs {
		matches := matcher.Resolve(input)
		output := make([]string, 0, len(matches))
		for _, match := range matches {
			if match.ClassroomKey != "" {
				output = append(output, match.ClassroomKey)
			} else {
				output = append(output, match.Input)
			}
		}
		fmt.Printf("%s -> %s\n", input, strings.Join(output, "/"))
	}
}
