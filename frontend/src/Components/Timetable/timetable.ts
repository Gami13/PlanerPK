import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

export const weekdays = [
	{ key: "Monday", label: "Poniedziałek" },
	{ key: "Tuesday", label: "Wtorek" },
	{ key: "Wednesday", label: "Środa" },
	{ key: "Thursday", label: "Czwartek" },
	{ key: "Friday", label: "Piątek" },
] as const;

export const eventPaletteNames = [
	"blue",
	"green",
	"amber",
	"violet",
	"teal",
	"rose",
	"indigo",
	"cyan",
	"olive",
	"sand",
] as const;

export type EventPalette = (typeof eventPaletteNames)[number];

// Stable per-course color: the same subject keeps its color across days/weeks,
// which makes the grid scannable without a legend.
export function paletteFor(seed: string): EventPalette {
	let hash = 0;
	for (let index = 0; index < seed.length; index++) {
		hash = (hash * 31 + seed.charCodeAt(index)) >>> 0;
	}
	return eventPaletteNames[hash % eventPaletteNames.length];
}

// "co tydz." / "co tydzień" is the norm, so it only adds noise. Anything else
// (e.g. "co 2 tydzień - 15h") is real information worth showing.
export function isWeeklyRecurrence(value: string | undefined): boolean {
	const normalized = value?.trim().toLocaleLowerCase("pl") ?? "";
	return normalized === "" || /^co\s+(?:1\s+)?tydz/.test(normalized);
}

export type TimetableEntry = {
	id: string;
	classInfo: ClassInfo;
	start: number;
	end: number;
	ghost: boolean;
};

export type PositionedEntry = TimetableEntry & { lane: number; lanes: number };

export function parseTime(value: string): number | undefined {
	const match = /^(\d{1,2}):(\d{2})(?::\d{2})?$/.exec(value);
	if (!match) return undefined;
	const hours = Number(match[1]);
	const minutes = Number(match[2]);
	return hours < 24 && minutes < 60 ? hours * 60 + minutes : undefined;
}

export function formatTime(minutes: number): string {
	return `${String(Math.floor(minutes / 60)).padStart(2, "0")}:${String(minutes % 60).padStart(2, "0")}`;
}

export function prepareEntries(classes: ClassInfo[], overrides?: Map<string, ClassOverride>): TimetableEntry[] {
	return classes.flatMap((classInfo) => {
		if (!weekdays.some((day) => day.key === classInfo.day)) return [];
		const start = parseTime(classInfo.startTime);
		const end = parseTime(classInfo.endTime);
		if (start === undefined || end === undefined || end <= start) return [];
		const override = overrides?.get(classInfo.hash);
		const shifted = parseTime(override?.startHour ?? "");
		const entries: TimetableEntry[] = [];
		if (!override?.isHidden) {
			const effectiveStart = shifted ?? start;
			entries.push({
				id: classInfo.hash,
				classInfo,
				start: effectiveStart,
				end: effectiveStart + end - start,
				ghost: false,
			});
		}
		// Match the calendar resolver: a transparent original exists only for a hidden or moved class.
		if (override?.isTransparent && (override.isHidden || shifted !== undefined)) {
			entries.push({ id: `${classInfo.hash}-original`, classInfo, start, end, ghost: true });
		}
		return entries;
	});
}

// Assign lanes per connected overlap group, so unrelated classes retain the full column width.
export function positionEntries(entries: TimetableEntry[]): PositionedEntry[] {
	const sorted = [...entries].sort((a, b) => a.start - b.start || b.end - a.end || a.id.localeCompare(b.id));
	const result: PositionedEntry[] = [];
	let group: PositionedEntry[] = [];
	let laneEnds: number[] = [];
	let groupEnd = -1;
	const flush = () => {
		for (const entry of group) result.push({ ...entry, lanes: laneEnds.length });
		group = [];
		laneEnds = [];
	};
	for (const entry of sorted) {
		if (entry.start >= groupEnd) flush();
		let lane = laneEnds.findIndex((end) => end <= entry.start);
		if (lane === -1) lane = laneEnds.length;
		laneEnds[lane] = entry.end;
		group.push({ ...entry, lane, lanes: 1 });
		groupEnd = Math.max(groupEnd, entry.end);
	}
	flush();
	return result;
}
