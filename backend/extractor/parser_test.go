package extractor

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gami13/PlanerPK/types"
)

// mockLegend returns a Legend populated with sample teachers and locations
// based on real timetable data, for use across all parser tests.
func mockLegend() *types.Legend {
	return &types.Legend{
		Teachers: map[string]types.Teacher{
			"JD":     {Shorthand: "JD", FullName: "Jan Doe"},
			"MD":     {Shorthand: "MD", FullName: "Michał Drygaś"},
			"TG":     {Shorthand: "TG", FullName: "T. Gąciarz"},
			"KSw":    {Shorthand: "KSw", FullName: "K. Świszcz"},
			"SF":     {Shorthand: "SF", FullName: "S. Fialko"},
			"AJ-S":   {Shorthand: "AJ-S", FullName: "A. Jasińska-Suwada"},
			"LB":     {Shorthand: "LB", FullName: "L. Bieniasz"},
			"MB":     {Shorthand: "MB", FullName: "M. Bereta"},
			"MN":     {Shorthand: "MN", FullName: "M. Niedźwiecki"},
			"JO":     {Shorthand: "JO", FullName: "J. Opiła"},
			"AP":     {Shorthand: "AP", FullName: "A. Plichta"},
			"OB":     {Shorthand: "OB", FullName: "O. Bar"},
			"KW":     {Shorthand: "KW", FullName: "K. Wiśniewski"},
			"FT":     {Shorthand: "FT", FullName: "F. Turza"},
			"WK":     {Shorthand: "WK", FullName: "W. Książek"},
			"ARP":    {Shorthand: "ARP", FullName: "A. R. P."},
			"DŻ":     {Shorthand: "DŻ", FullName: "D. Żelasko"},
			"TL":     {Shorthand: "TL", FullName: "T. L."},
			"MNiedz": {Shorthand: "MNiedz", FullName: "M. Niedźwiecki"},
			"MNaw":   {Shorthand: "MNaw", FullName: "M. Nawrocki"},
			"DK":     {Shorthand: "DK", FullName: "D. Kulis"},
			"SzSzom": {Shorthand: "SzSzom", FullName: "Szymon Szomiński"},
			"BB":     {Shorthand: "BB", FullName: "B. B."},
			"KS":     {Shorthand: "KS", FullName: "K. Skabek"},
			"IU":     {Shorthand: "IU", FullName: "I. U."},
			"FK":     {Shorthand: "FK", FullName: "F. K."},
			"RK":     {Shorthand: "RK", FullName: "R. K."},
			"JK":     {Shorthand: "JK", FullName: "J. K."},
			"AWid":   {Shorthand: "AWid", FullName: "A. Wid"},
			"AW":     {Shorthand: "AW", FullName: "A. W."},
			"EPAM":   {Shorthand: "EPAM", FullName: "EPAM"},
			"NOKIA":  {Shorthand: "NOKIA", FullName: "NOKIA"},
		},
		Locations: map[string]types.Location{
			"131":                       {Shorthand: "131", FullName: "s. 131"},
			"135":                       {Shorthand: "135", FullName: "s. 135"},
			"136":                       {Shorthand: "136", FullName: "s. 136"},
			"151":                       {Shorthand: "151", FullName: "s. 151"},
			"114 GIL":                   {Shorthand: "114 GIL", FullName: "s. 114 GIL"},
			"152 SJO":                   {Shorthand: "152 SJO", FullName: "s. 152 SJO"},
			"N":                         {Shorthand: "N", FullName: "s. N"},
			"SJO":                       {Shorthand: "SJO", FullName: "s. SJO"},
			"S1":                        {Shorthand: "S1", FullName: "s. S1"},
			"F020":                      {Shorthand: "F020", FullName: "s. F020"},
			"Działownia 1/15":           {Shorthand: "Działownia 1/15", FullName: "Działownia 1/15"},
			"wykładowa KI/135, 136 bis": {Shorthand: "wykładowa KI/135, 136 bis", FullName: "Wykładowa KI"},
		},
	}
}

var (
	testDay       = types.Day{StartingRow: 0, EndingRow: 10, WeekDay: time.Monday}
	testStartTime = time.Date(2026, 1, 5, 8, 0, 0, 0, time.UTC)
	testEndTime   = time.Date(2026, 1, 5, 9, 30, 0, 0, time.UTC)
)

func TestParseClassText(t *testing.T) {
	legend := mockLegend()

	tests := []struct {
		name       string
		input      string
		nilLegend  bool
		want       types.Class
		extraCheck func(t *testing.T, got types.Class)
	}{
		{
			name:  "BasicLecture",
			input: "Analiza matematyczna  \nwykład\ndr J. Pudełko     ZDALNIE     ",
			want:  types.Class{Name: "Analiza matematyczna", Type: "wykład", Teacher: "dr J. Pudełko", Location: "ZDALNIE"},
		},
		{
			name:  "LectureWithFrequencyAndHours",
			input: "Programowanie w języku Python     wykład\ndr inż. T. Gąciarz     co 2 tydzień - 15h     ZDALNIE     ",
			want:  types.Class{Name: "Programowanie w języku Python", Type: "wykład", Teacher: "dr inż. T. Gąciarz", Location: "ZDALNIE", AdditionalInfo: "co 2 tydzień"},
		},
		{
			name:  "LabWithShorthandTeacher",
			input: "PEiTC     lab.     KSw     s. 131",
			want:  types.Class{Name: "PEiTC", Type: "lab.", Teacher: "KSw", Location: "s. 131"},
		},
		{
			name:  "LanguageWithLevel",
			input: "J. niemiecki/J.angielski C1\nćw.\nmgr D. Guzik/mgr E. Szabat     s. SJO/s. N     ",
			want:  types.Class{Name: "J. niemiecki/J.angielski", Type: "ćw.", Level: "C1", Teacher: "mgr D. Guzik/mgr E. Szabat", Location: "s. SJO/s. N"},
		},
		{
			name:  "ExerciseWithFullLocation",
			input: "Analiza matematyczna\nćw.\nmgr M. Ból     s. Działownia 1/15",
			want:  types.Class{Name: "Analiza matematyczna", Type: "ćw.", Teacher: "mgr M. Ból", Location: "s. Działownia 1/15"},
		},
		{
			name:  "SlashDelimitedTeachers",
			input: "AiSD     lab.     JD/MD     s. 131",
			want:  types.Class{Name: "AiSD", Type: "lab.", Teacher: "JD/MD", Location: "s. 131"},
		},
		{
			name:  "ProjectWithHoursAndWarning",
			input: "JiPP\nP     JD/MD     (15h)(5h/10h)     s. 131     UWAGA! Zajęcia rozpoczynają się w drugiej połowie semestru.",
			want:  types.Class{Name: "JiPP", Type: "P", Teacher: "JD/MD", Location: "s. 131", TotalHours: "(15h)(5h/10h)", AdditionalInfo: "UWAGA!"},
		},
		{
			name:  "LabMergedWithTeacher",
			input: "PEiTC     lab. Michał Drygaś     s. 131",
			want:  types.Class{Name: "PEiTC", Type: "lab.", Location: "s. 131", AdditionalInfo: "Michał Drygaś"},
		},
		{
			name:  "MultiTeacherWithHours",
			input: "GK     lab. MN/JO/W. Malik/G. Kuliś    (5h/5h/10h/10h)     s. 135",
			want:  types.Class{Name: "GK", Type: "lab.", Teacher: "MN/JO/W. Malik/G. Kuliś", TotalHours: "(5h/5h/10h/10h)", Location: "s. 135"},
		},
		{
			name:  "MixedLocationZdalnie",
			input: "PUS     P     MNiedz     s. 135/ZDALNIE",
			want:  types.Class{Name: "PUS", Type: "P", Teacher: "MNiedz", Location: "s. 135/ZDALNIE"},
		},
		{
			name:  "FullLectureWithProfTitle",
			input: "Języki i paradygmaty programowania\nwykład\ndr hab. inż. S. Fialko, prof. PK     ZDALNIE     ",
			want:  types.Class{Name: "Języki i paradygmaty programowania", Type: "wykład", Teacher: "dr hab. inż. S. Fialko, prof. PK", Location: "ZDALNIE"},
		},
		{
			name:  "ExerciseLabCombo",
			input: "Metody obliczeniowe\nćw./lab. \nLB     co 2 tydzień - 15h     s. N/s.151",
			want:  types.Class{Name: "Metody obliczeniowe", Type: "ćw./lab.", Teacher: "LB", AdditionalInfo: "co 2 tydzień", Location: "s. N/s.151"},
		},
		{
			name:  "RPiSLabWithFullTeacherName",
			input: "RPiS     lab.     E. Gajecka-Mirek     s. 135",
			want:  types.Class{Name: "RPiS", Type: "lab.", Location: "s. 135", AdditionalInfo: "E. Gajecka-Mirek"},
		},
		{
			name:  "LabWithComplexLocation",
			input: "RPiS     lab.     E. Gajecka-Mirek     s. wykładowa KI/135, 136 bis",
			want:  types.Class{Name: "RPiS", Type: "lab.", Location: "s. wykładowa KI/135, 136 bis"},
		},
		{
			name:  "BDIILabWithHoursInTeacher",
			input: "BD II\nlab.     AP/OB/SzSzom (14h/8h/8h)     s. 136",
			want:  types.Class{Name: "BD II", Type: "lab.", Location: "s. 136"},
		},
		{
			name:  "LanguageLevelInSecondPart",
			input: "J.angielski \n C1     ćw.\nmgr E. Szabat     s. N     ",
			want:  types.Class{Name: "J.angielski", Type: "ćw.", Teacher: "mgr E. Szabat"},
		},
		{
			name:  "ProjectWithGroupInfo",
			input: "PTS     lab. gr 21/1/ćw. gr 21     FT     s. 114 GIL/Działownia 1/15     UWAGA!     Laboratoria rozpoczynają się od początku semestru.     Ćwiczenia rozpoczynają się w połowie semestru.",
			want:  types.Class{Name: "PTS", Type: "lab.", Teacher: "FT", Location: "s. 114 GIL/Działownia 1/15", AdditionalInfo: "UWAGA!"},
		},
		{
			name:  "SeminarWithHours",
			input: "Przygotowanie dyplomowe seminarium     FT     s. S1     (15h)",
			want:  types.Class{Name: "Przygotowanie dyplomowe seminarium", Teacher: "FT", Location: "s. S1", TotalHours: "(15h)"},
		},
		{
			name:  "NetLabWithFullName",
			input: ".NET     lab.     Łukasz Świszcz     s. 151",
			want:  types.Class{Name: ".NET", Type: "lab.", Location: "s. 151"},
		},
		{
			name:  "TeacherWithLocationNoType",
			input: "J. niemiecki mgr Dariusz Guzik     s. SJO",
			want:  types.Class{Name: "J. niemiecki mgr Dariusz Guzik", Location: "s. SJO"},
		},
		{
			name:  "LabWithShorthandAndFrequency",
			input: "MO\nlab.     LB     co 2 tydzień - 15h     s. 151",
			want:  types.Class{Name: "MO", Type: "lab.", Teacher: "LB", AdditionalInfo: "co 2 tydzień", Location: "s. 151"},
		},
		{
			name:  "WSILabWithHours",
			input: "WSI\nlab.     MB     (30h)     s. 136",
			want:  types.Class{Name: "WSI", Type: "lab.", Teacher: "MB", TotalHours: "(30h)", Location: "s. 136"},
		},
		{
			name:  "MultiLectureWithSlashes",
			input: "Bazy danych II     \nwykład     dr inż.  A. Plichta/dr O. Bar/Szymon Szomiński     (14h/8h/8h)     ZDALNIE",
			want:  types.Class{Name: "Bazy danych II", Type: "wykład", Teacher: "dr inż. A. Plichta/dr O. Bar/Szymon Szomiński", TotalHours: "(14h/8h/8h)", Location: "ZDALNIE"},
		},
		{
			name:  "LabWithBisLocation",
			input: "WSI\nlab.     MB/ARP (12h/18h)     s. wykładowa KI/135, 136 bis",
			want:  types.Class{Name: "WSI", Type: "lab.", Location: "s. wykładowa KI/135, 136 bis"},
		},
		{
			name:  "HighPerformanceComputing",
			input: "High Performance Computing\nFK     s. 135",
			want:  types.Class{Name: "High Performance Computing", Teacher: "FK", Location: "s. 135"},
		},
		{
			name:  "JavaLabZdalnie",
			input: "Java\nlab.     EPAM     ZDALNIE",
			want:  types.Class{Name: "Java", Type: "lab.", Teacher: "EPAM", Location: "ZDALNIE"},
		},
		{
			name:  "MobileWithWymiennie",
			input: "Mobile Technologies and Programming     AW     s. 114 GIL     (lab. 15h wymiennie co 2 tydzień)     ",
			want:  types.Class{Name: "Mobile Technologies and Programming", Teacher: "AW", Location: "s. 114 GIL", AdditionalInfo: "wymiennie"},
		},
		{
			name:  "NameOnly",
			input: "Wychowanie fizyczne     ",
			want:  types.Class{Name: "Wychowanie fizyczne"},
		},
		{
			name:  "EmptyString",
			input: "",
			want:  types.Class{},
		},
		{
			name:      "NilLegend",
			input:     "AiSD     lab.     JD     s. 131",
			nilLegend: true,
			want:      types.Class{Name: "AiSD", Type: "lab.", Location: "s. 131"},
			extraCheck: func(t *testing.T, got types.Class) {
				if got.Teacher == "" && got.AdditionalInfo == "" {
					t.Error("expected 'JD' in Teacher or AdditionalInfo, both are empty")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lg := legend
			if tt.nilLegend {
				lg = nil
			}
			class := parseClassText(tt.input, testStartTime, testEndTime, testDay, lg)
			assertStructContains(t, class, tt.want)
			if tt.extraCheck != nil {
				tt.extraCheck(t, class)
			}
		})
	}
}

func TestSplitSlashClasses(t *testing.T) {
	legend := mockLegend()

	tests := []struct {
		name  string
		input string
		want  []types.Class
	}{
		{
			name:  "LanguageSplitWithLevel",
			input: "J. niemiecki/J.angielski C1\nćw.\nmgr D. Guzik/mgr E. Szabat     s. SJO/s. N     ",
			want: []types.Class{
				{Name: "J. niemiecki", Type: "ćw.", Teacher: "mgr D. Guzik", Location: "s. SJO"},
				{Name: "J.angielski", Type: "ćw.", Level: "C1", Teacher: "mgr E. Szabat", Location: "s. N"},
			},
		},
		{
			name:  "NoSlashInNameNoSplit",
			input: "AiSD     lab.     JD/MD     s. 131",
			want: []types.Class{
				{Name: "AiSD", Type: "lab.", Teacher: "JD/MD", Location: "s. 131"},
			},
		},
		{
			name:  "SlashNameMismatchedFieldCounts",
			input: "Subject A/Subject B\nlab.     JD     s. 131",
			want: []types.Class{
				{Name: "Subject A", Type: "lab.", Teacher: "JD", Location: "s. 131"},
				{Name: "Subject B", Type: "lab.", Teacher: "JD", Location: "s. 131"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed := parseClassText(tt.input, testStartTime, testEndTime, testDay, legend)
			got := splitSlashClasses(parsed)
			if len(got) != len(tt.want) {
				t.Fatalf("got %d classes, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				assertStructContains(t, got[i], tt.want[i])
			}
		})
	}
}

// --- helpers ---

func assertStructContains(t *testing.T, got, want types.Class) {
	t.Helper()
	gotV := reflect.ValueOf(got)
	wantV := reflect.ValueOf(want)
	typ := wantV.Type()
	for i := 0; i < typ.NumField(); i++ {
		wantStr, ok := wantV.Field(i).Interface().(string)
		if !ok || wantStr == "" {
			continue
		}
		gotStr := gotV.Field(i).Interface().(string)
		if !strings.Contains(gotStr, wantStr) {
			t.Errorf("%s: got %q, want it to contain %q", typ.Field(i).Name, gotStr, wantStr)
		}
	}
}
