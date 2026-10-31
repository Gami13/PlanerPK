export const guestTimetableKey = "planerpk.previewColumn";

export function readGuestSelection(): number | undefined {
	try {
		const value = localStorage.getItem(guestTimetableKey);
		if (value === null || !/^\d+$/.test(value)) return undefined;
		const column = Number(value);
		return Number.isSafeInteger(column) ? column : undefined;
	} catch {
		return undefined;
	}
}

export function saveGuestSelection(column: number): void {
	try {
		localStorage.setItem(guestTimetableKey, String(column));
	} catch {
		// Browsing still works when storage is unavailable.
	}
}

export function resolveGuestSelection(columns: readonly { column: number }[], saved?: number): number | undefined {
	return columns.find((column) => column.column === saved)?.column ?? columns[0]?.column;
}
