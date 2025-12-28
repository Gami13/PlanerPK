package global

var (
	SHEET_NAME                     = "2025_2026"
	LAST_ROW                       = 365
	HOUR_COLUMN                    = 1
	DAY_START                      = "7.30"
	DAY_END                        = "21.00"
	DATABASE_SOURCE                = "file:../data/database.db?cache=shared&mode=rwc&_busy_timeout=5000"
	LAST_COLUMN                    = 40
	DATABASE_DRIVER                = "sqlite"
	MINUTES_BETWEEN                = 5
	SERVER_PORT                    = "8080"
	GENERAL_RATE_LIMIT_PER_MINUTE  = 60
	TIMETABLE_CURRENT_DESTINATION  = "../data/current_timetable/"
	TIMETABLE_URL                  = "https://it.pk.edu.pl/studenci/na-studiach/rozklady-zajec/"
	TIMETABLE_DOWNLOAD_DESTINATION = "../data/downloads/"
	CALENDAR_CREDENTIALS_PATH      = "../credentials.json"
	LOGS_PATH                      = "../data/logs"
	TIME_BETWEEN_SYNC_HOURS        = 72
)
