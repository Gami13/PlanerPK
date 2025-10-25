package types

type Config struct {
	SheetName  string
	LastRow    int
	HourColumn int
	DayStart   string
	DayEnd     string
}

func NewConfig() *Config {
	return &Config{
		SheetName:  "2025_2026",
		LastRow:    365,
		HourColumn: 1,
		DayStart:   "7.30",
		DayEnd:     "21.00",
	}
}
