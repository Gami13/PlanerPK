export const CONTACT_EMAIL = "gamiwastaken@gmail.com";
export const PK_WEBSITE_URL = "https://www.pk.edu.pl/index.php?lang=pl";
export const GOOGLE_CALENDAR_URL = "https://calendar.google.com/calendar?hl=pl";

export const DayLabels: Record<string, string> = {
	Monday: "Pon",
	Tuesday: "Wt",
	Wednesday: "Śr",
	Thursday: "Czw",
	Friday: "Pt",
	Saturday: "Sob",
	Sunday: "Ndz",
};

export const UIText = {
	settings: {
		account: {
			title: "Synchronizacja",
			statusError: "Wystąpił błąd podczas zmiany statusu konta. Spróbuj ponownie.",
		},
		column: {
			title: "Wybrana kolumna",
			loadingColumns: "Ładowanie kolumn",
			errorSetting: "Wystąpił błąd podczas zmiany kolumny. Spróbuj ponownie.",
			successSetting: "Pomyślnie zmieniono kolumnę.",
			notSelected: "Nie wybrano kolumny",
			change: "Zmień kolumnę",
		},
	},
} as const;
