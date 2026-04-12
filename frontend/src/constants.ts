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
export const UIText = {
	// Status messages
	status: {
		loading: "Ładowanie...",
	},
	// Account status

	settings: {
		account: {
			title: "Status aktywności",
			statusError: "Wystąpił błąd podczas zmiany statusu konta. Spróbuj ponownie.",
			active: "Aktywny",
			inactive: "Nieaktywny",
			markActive: "Oznacz konto jako aktywne",
			markInactive: "Oznacz konto jako nieaktywne",
		},
		column: {
			title: "Wybrana kolumna",
			description: "Zmianę kolumny wykonuj wyłącznie w uzasadnionych przypadkach, aby uniknąć ograniczeń konta.",
			loadingColumns: "Ładowanie kolumn",
			errorLoadingColumns: "Błąd ładowania kolumn",
			errorSetting: "Wystąpił błąd podczas zmiany kolumny. Spróbuj ponownie.",
			successSetting: "Pomyślnie zmieniono kolumnę.",
			notSelected: "Nie wybrano kolumny",
			change: "Zmień kolumnę",
		},
		classOverrides: {
			title: "Nadpisania zajęć",
			description:
				"Możesz ukryć zajęcia, przesunąć ich godzinę lub zostawić poprzednią pozycję jako transparentną w kalendarzu. Zmiany zostaną zastosowane przy następnej synchronizacji.",
			hideClass: "Ukryj zajęcia",
			showTransparent: "Pokaż poprzednią pozycję jako transparentną",
			shiftTime: "Przesuń na godzinę",
			saving: "Zapisywanie…",
			save: "Zapisz",
			saved: "✓ Zapisano",
			saveError: "✗ Błąd zapisu",
			loadingClasses: "Ładowanie zajęć…",
			errorLoadingClasses: "Błąd ładowania zajęć.",
		},
	},
} as const;
