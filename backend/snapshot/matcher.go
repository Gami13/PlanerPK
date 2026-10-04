package snapshot

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/woodpkrs/PlanerPK/types"
)

var audienceGroup = regexp.MustCompile(`^(?:GL|C|gr\. |grupa obieralna )([0-9]+)(?: \(GL[0-9]+(?:\+GL[0-9]+)*\))?$`)
var cohortGroup = regexp.MustCompile(` / gr\. ([0-9]+)$`)
var mastersExerciseAudience = regexp.MustCompile(`^(?:CY|DS|SIRR)_C[0-9]+$`)

// Candidates tries the exact subject first, then other subjects if no candidates
// remain. The fallback requires a known matching teacher. Multiple results stay
// ambiguous; callers must not choose one arbitrarily for scheduling.
func (index SubjectIndex) Candidates(class types.Class) []Block {
	matches := index.match(class, index[class.Name], false)
	if len(matches) != 0 || strings.TrimSpace(class.Teacher) == "" {
		return matches
	}
	var blocks []Block
	for _, subjectBlocks := range index {
		blocks = append(blocks, subjectBlocks...)
	}
	matches = index.match(class, blocks, true)
	sort.Slice(matches, func(i, j int) bool { return matches[i].ID < matches[j].ID })
	return matches
}

func (index SubjectIndex) match(class types.Class, blocks []Block, requireTeacher bool) []Block {
	days := map[time.Weekday]string{
		time.Monday: "MON", time.Tuesday: "TUE", time.Wednesday: "WED",
		time.Thursday: "THU", time.Friday: "FRI",
	}
	day, ok := days[class.Day]
	if !ok {
		return nil
	}
	start := class.StartTime.Hour()*60 + class.StartTime.Minute()
	teacher := strings.TrimSpace(class.Teacher)
	note, _, _ := strings.Cut(strings.TrimSpace(class.AdditionalInfo), "\n")
	note = strings.TrimSpace(note)
	date, dateErr := time.Parse("02.01.2006", note)
	var matches []Block
	for _, block := range blocks {
		if block.Activity != class.Type || block.Day == nil || *block.Day != day || block.Start == nil || *block.Start != start {
			continue
		}
		if !block.matchesAudience(class.Audience) {
			continue
		}
		if (note == "A" || note == "B") && block.Frequency == "co_2_tygodnie" && block.TeachingWeekParity != nil {
			parity := 0
			if note == "B" {
				parity = 1
			}
			if *block.TeachingWeekParity != parity {
				continue
			}
		}
		if dateErr == nil && block.Date != nil && *block.Date != date.Format("2006-01-02") {
			continue
		}
		name, display := strings.TrimSpace(block.Teacher), strings.TrimSpace(block.TeacherDisplay)
		if requireTeacher && teacher != name && teacher != display {
			continue
		}
		if teacher != "" && (name != "" || display != "") && teacher != name && teacher != display {
			continue
		}
		matches = append(matches, block)
	}
	return matches
}

func (block Block) matchesAudience(audience string) bool {
	if mastersExerciseAudience.MatchString(audience) && block.AcademicGroupLabel != "" {
		return block.AcademicGroupLabel == audience
	}
	groupNumber := audienceGroup.FindStringSubmatch(audience)
	if len(groupNumber) == 0 {
		return true
	}
	group := ""
	if block.GroupNo != nil {
		group = strconv.Itoa(*block.GroupNo)
	} else if cohort := cohortGroup.FindStringSubmatch(block.Cohort); len(cohort) != 0 {
		group = cohort[1]
	}
	return group == "" || group == groupNumber[1]
}
