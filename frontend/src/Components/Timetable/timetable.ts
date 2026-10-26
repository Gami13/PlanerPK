import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

export const weekdays = [
	{ key: "Monday", label: "Poniedziałek" },
	{ key: "Tuesday", label: "Wtorek" },
	{ key: "Wednesday", label: "Środa" },
	{ key: "Thursday", label: "Czwartek" },
	{ key: "Friday", label: "Piątek" },
] as const;

// "co tydz." / "co tydzień" is the norm, so it only adds noise. Anything else
// (e.g. "co 2 tydzień - 15h") is real information worth showing.
export function isWeeklyRecurrence(value: string | undefined): boolean {
	const normalized = value?.trim().toLocaleLowerCase("pl") ?? "";
	return normalized === "" || /^co\s+(?:1\s+)?tydz/.test(normalized);
}

export function weekMarker(value: string | undefined): string | undefined {
	const match = /^\s*(?:tydzień\s+)?([AB])\s*$/i.exec(value ?? "");
	return match?.[1].toUpperCase();
}

function labGroups(value: string): string[] {
	return [...value.toUpperCase().matchAll(/\bG[LK](\d+)\b/g)].map((match) => `GL${Number(match[1])}`);
}

export function shouldShowAudience(type: string, audience: string, selectedGroup?: string): boolean {
	if (!audience.trim() || /^(w|wykład|wyklad|wyk\.?)$/i.test(type.trim())) return false;
	if (!selectedGroup) return true;
	const normalize = (value: string) => value.trim().replace(/\s+/g, " ").toUpperCase();
	if (normalize(audience) === normalize(selectedGroup)) return false;
	const selectedLabs = labGroups(selectedGroup);
	const audienceLabs = labGroups(audience);
	// Explicit membership such as C1 (GL1+GL2) takes priority over numbering.
	if (selectedLabs.length && audienceLabs.length) {
		return !selectedLabs.every((group) => audienceLabs.includes(group));
	}
	// GC1 / C1 is the parent of GL1 and GL2 (also named GK1 and GK2).
	const parent = /^(?:GC|C)(\d+)$/i.exec(audience.trim());
	if (parent && selectedLabs.length) {
		return !selectedLabs.every((group) => Math.ceil(Number(group.slice(2)) / 2) === Number(parent[1]));
	}
	return true;
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
