export const PK_WEBSITE_URL = "https://www.pk.edu.pl/index.php?lang=pl";
export const GOOGLE_CALENDAR_URL = "https://calendar.google.com/calendar?hl=pl";

export const DAY_LABELS: Record<string, string> = {
	Monday: "Pon",
	Tuesday: "Wt",
	Wednesday: "Śr",
	Thursday: "Czw",
	Friday: "Pt",
	Saturday: "Sob",
	Sunday: "Ndz",
};

// UI Messages - Settings
export const UI_MESSAGES = {
	// Status messages
	LOADING: "Ładowanie...",
	LOADING_CLASSES: "Ładowanie zajęć…",
	LOADING_COLUMNS: "Ładowanie kolumn",
	ERROR_LOADING_CLASSES: "Błąd ładowania zajęć.",
	ERROR_LOADING_COLUMNS: "Błąd ładowania kolumn",

	// Account status
	ACCOUNT_STATUS_ERROR: "Wystąpił błąd podczas zmiany statusu konta. Spróbuj ponownie.",
	ACCOUNT_ACTIVE: "Aktywny",
	ACCOUNT_INACTIVE: "Nieaktywny",
	MARK_ACTIVE: "Oznacz konto jako aktywne",
	MARK_INACTIVE: "Oznacz konto jako nieaktywne",

	// Column selection
	COLUMN_ERROR: "Wystąpił błąd podczas zmiany kolumny. Spróbuj ponownie.",
	COLUMN_SUCCESS: "Pomyślnie zmieniono kolumnę.",
	COLUMN_NOT_SELECTED: "Nie wybrano kolumny",
	CHANGE_COLUMN: "Zmień kolumnę",

	// Class overrides
	HIDE_CLASS: "Ukryj zajęcia",
	SHOW_TRANSPARENT: "Pokaż poprzednią pozycję jako transparentną",
	SHIFT_TIME: "Przesuń na godzinę",
	SAVING: "Zapisywanie…",
	SAVE: "Zapisz",
	SAVED: "✓ Zapisano",
	SAVE_ERROR: "✗ Błąd zapisu",
} as const;
