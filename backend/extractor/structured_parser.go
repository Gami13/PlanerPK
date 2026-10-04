package extractor

import (
	"regexp"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/internal/clock"
	"github.com/woodpkrs/PlanerPK/types"
)

var (
	blankLines         = regexp.MustCompile(`\n(?:[\t ]*\n)+`)
	structuredTypeCode = regexp.MustCompile(`^(?:W|L|C|P|WF)(?:[+/](?:W|L|C|P|WF))*$`)
)

// parseStructuredClassText parses every blank-line-delimited fixed-field
// record. Invalid records are ignored instead of being partially interpreted.
func parseStructuredClassText(text string, day types.Day) []types.Class {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	var classes []types.Class
	for _, record := range blankLines.Split(strings.TrimSpace(text), -1) {
		if class, ok := parseStructuredRecord(record, day); ok {
			classes = append(classes, class)
		}
	}
	return classes
}

func parseStructuredRecord(text string, day types.Day) (types.Class, bool) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	for i := range lines {
		lines[i] = strings.TrimSpace(lines[i])
	}
	if len(lines) < 4 || lines[1] == "" || lines[2] == "" {
		return types.Class{}, false
	}

	name, classType, ok := parseStructuredHeader(lines[0])
	if !ok {
		return types.Class{}, false
	}
	start, end, location, ok := parseTimeLocation(lines[3], day.StartingTime)
	if !ok {
		return types.Class{}, false
	}

	return types.Class{
		Name:           name,
		Type:           classType,
		Day:            day.WeekDay,
		StartTime:      start,
		EndTime:        end,
		Teacher:        lines[1],
		Audience:       lines[2],
		Location:       location,
		AdditionalInfo: strings.Join(lines[4:], "\n"),
	}, true
}

func parseStructuredHeader(line string) (name, classType string, ok bool) {
	index := strings.LastIndexAny(line, "—–-")
	if index < 0 {
		return "", "", false
	}
	name = strings.TrimSpace(line[:index])
	classType = strings.TrimSpace(strings.TrimLeft(line[index:], "—–-"))
	return name, classType, name != "" && structuredTypeCode.MatchString(classType)
}

func parseTimeLocation(line string, base time.Time) (start, end time.Time, location string, ok bool) {
	rangeText, location, ok := strings.Cut(line, "·")
	if !ok || strings.TrimSpace(location) == "" {
		return time.Time{}, time.Time{}, "", false
	}
	times := strings.FieldsFunc(rangeText, func(r rune) bool { return r == '–' || r == '-' })
	if len(times) != 2 {
		return time.Time{}, time.Time{}, "", false
	}
	start, startOK := timeOnDate(strings.TrimSpace(times[0]), base)
	end, endOK := timeOnDate(strings.TrimSpace(times[1]), base)
	return start, end, strings.TrimSpace(location), startOK && endOK && end.After(start)
}

func timeOnDate(value string, date time.Time) (time.Time, bool) {
	clockTime, ok := clock.Parse(value)
	if !ok {
		return time.Time{}, false
	}
	return time.Date(date.Year(), date.Month(), date.Day(), clockTime.Hour(), clockTime.Minute(), 0, 0, date.Location()), true
}
