package utils

import (
	"time"
)

func GetDateOfNextWeekday(weekday time.Weekday) time.Time {
	today := time.Now()
	daysUntil := (int(weekday) - int(today.Weekday()) + 7) % 7
	next := today.AddDate(0, 0, daysUntil)
	return time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, time.UTC)
}

//JiPP - Języki i paradygmaty programowania
//WdI - Wstęp do informatyki
//F - Fizyka
//AM - Analiza matematyczna
//J.angielski - Język angielski
//LdI - Logika dla inżynierów
//AzG - Algebra z geometrią
//J. niemiecki - Język niemiecki

//wykład, lab., ćwiczenia
