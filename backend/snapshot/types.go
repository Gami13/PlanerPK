// Package snapshot contains the wire types for the published timetable snapshot.
// These types describe source data, not the application's selected groups or events.
package snapshot

// Response is the snapshot HTTP envelope. CalendarExport is only an offline
// test oracle; production timetables must be derived from State.
type Response struct {
	State          State           `json:"state"`
	CalendarExport *CalendarExport `json:"calendarExport"`
}

type State struct {
	SchemaVersion    int              `json:"schemaVersion"`
	PublicationMode  string           `json:"publicationMode"`
	Blocks           []Block          `json:"blocks"`
	Rooms            []Room           `json:"rooms"`
	AcademicCalendar AcademicCalendar `json:"academicCalendar"`
	Meta             StateMeta        `json:"meta"`
	Snapshot         SnapshotMeta     `json:"snapshot"`
}

type Block struct {
	ID                 string           `json:"id"`
	Subject            string           `json:"subject"`
	Activity           string           `json:"activity"`
	Teacher            string           `json:"teacher"`
	TeacherDisplay     string           `json:"teacherDisplay"`
	Cohort             string           `json:"cohort"`
	AdditionalCohorts  []string         `json:"additionalCohorts"`
	PlanType           string           `json:"planType"`
	Day                *string          `json:"day"`      // MON..FRI; nil for parked blocks.
	Start              *int             `json:"start"`    // Minutes since local midnight.
	Duration           int              `json:"duration"` // Minutes.
	Room               *string          `json:"room"`
	Campus             *string          `json:"campus"`
	Modality           string           `json:"modality"`
	Hours              *int             `json:"hours"`
	Frequency          string           `json:"frequency"`
	Period             string           `json:"period"`
	Date               *string          `json:"date"` // YYYY-MM-DD; nil when not explicitly dated.
	TeachingWeekParity *int             `json:"teachingWeekParity"`
	GroupNo            *int             `json:"groupNo"`
	StudentGrouping    *StudentGrouping `json:"studentGrouping"`
	AllowedDates       []string         `json:"allowedDates"`
	ExcludedDates      []string         `json:"excludedDates"`
	AcademicGroupLabel string           `json:"academicGroupLabel"`
	SharedSpecialties  []string         `json:"sharedSpecialties"`
	ParkingReason      string           `json:"parkingReason"`
}

type StudentGrouping struct {
	Kind              string   `json:"kind"`
	Pool              string   `json:"pool"`
	Group             *int     `json:"group"`
	TotalGroups       *int     `json:"totalGroups"`
	Version           string   `json:"version"`
	GroupIDs          []string `json:"groupIds"`
	LabGroupIDs       []string `json:"labGroupIds"`
	Specialties       []string `json:"specialties"`
	LegacyExcessGroup bool     `json:"legacyExcessGroup"`
}

// Calendar dates stay as strings until explicit validation and Warsaw-local
// parsing; a missing date must not imply a scheduled occurrence.
type AcademicCalendar struct {
	Year                   string            `json:"year"`
	Season                 string            `json:"season"`
	Start                  string            `json:"start"`
	End                    string            `json:"end"`
	Closures               []DateRange       `json:"closures"`
	PartialClosures        []PartialClosure  `json:"partialClosures"`
	ReplacementDays        map[string]string `json:"replacementDays"`
	ParityAnchor           string            `json:"parityAnchor"`
	FirstWeek              string            `json:"firstWeek"`
	HalfBoundary           string            `json:"halfBoundary"`
	RegularStationaryStart string            `json:"regularStationaryStart"`
}

type DateRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type PartialClosure struct {
	Date  string `json:"date"`
	Start int    `json:"start"` // Minutes since local midnight.
	End   int    `json:"end"`
}

type Room struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	Campus   string `json:"campus"`
	RoomType string `json:"room_type"`
	Capacity int    `json:"capacity"`
	Active   int    `json:"active"` // 0 or 1 in the source JSON.
}

type StateMeta struct {
	Term          string `json:"term"`
	SourceRefresh string `json:"sourceRefresh"`
}

type SnapshotMeta struct {
	SavedAt string `json:"savedAt"`
	SHA256  string `json:"sha256"` // Publisher metadata, not the HTTP body fingerprint.
}

// CalendarExport must only be used for offline comparison tests, not imports.
type CalendarExport struct {
	Version  int           `json:"version"`
	Timezone string        `json:"timezone"`
	Events   []ExportEvent `json:"events"`
}

type ExportEvent struct {
	BlockID string `json:"blockId"`
	UID     string `json:"uid"`
	Start   string `json:"start"` // UTC YYYYMMDDTHHMMSSZ.
	End     string `json:"end"`
}
