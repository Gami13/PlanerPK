import assert from "node:assert/strict";
import { test } from "node:test";
import { parseTime, formatTime, prepareEntries, positionEntries } from "../src/Components/Timetable/timetable.ts";

const lesson = (patch = {}) => ({
	hash: "lesson",
	name: "Matematyka",
	day: "Monday",
	startTime: "08:00",
	endTime: "09:30",
	...patch,
});

test("time parsing validates values and accepts seconds", () => {
	assert.equal(parseTime("8:05"), 485);
	assert.equal(parseTime("08:05:00"), 485);
	assert.equal(parseTime("24:00"), undefined);
	assert.equal(parseTime("08:60"), undefined);
	assert.equal(parseTime(""), undefined);
	assert.equal(formatTime(485), "08:05");
});

test("weekends and invalid time ranges are excluded", () => {
	assert.equal(
		prepareEntries([lesson({ day: "Saturday" }), lesson({ day: "Sunday" }), lesson({ endTime: "07:00" })]).length,
		0
	);
});

test("moving a lesson retains duration and optionally its transparent original", () => {
	const entries = prepareEntries([lesson()], new Map([["lesson", { startHour: "10:15", isTransparent: true }]]));
	assert.deepEqual(
		entries.map(({ start, end, ghost }) => ({ start, end, ghost })),
		[
			{ start: 615, end: 705, ghost: false },
			{ start: 480, end: 570, ghost: true },
		]
	);
});

test("hidden lessons disappear unless a transparent original is requested", () => {
	assert.equal(prepareEntries([lesson()], new Map([["lesson", { isHidden: true }]])).length, 0);
	const entries = prepareEntries(
		[lesson()],
		new Map([["lesson", { isHidden: true, isTransparent: true, startHour: "11:00" }]])
	);
	assert.equal(entries.length, 1);
	assert.equal(entries[0].start, 480);
	assert.equal(entries[0].ghost, true);
});

test("transparency alone does not duplicate an unchanged lesson", () => {
	assert.equal(prepareEntries([lesson()], new Map([["lesson", { isTransparent: true }]])).length, 1);
});

test("overlaps share lanes, touching and unrelated lessons use full width", () => {
	const entries = positionEntries(
		prepareEntries([
			lesson({ hash: "a", startTime: "08:00", endTime: "09:30" }),
			lesson({ hash: "b", startTime: "08:30", endTime: "10:00" }),
			lesson({ hash: "c", startTime: "09:30", endTime: "10:00" }),
			lesson({ hash: "d", startTime: "10:00", endTime: "11:00" }),
		])
	);
	assert.deepEqual(
		entries.map(({ lane, lanes }) => [lane, lanes]),
		[
			[0, 2],
			[1, 2],
			[0, 2],
			[0, 1],
		]
	);
});
