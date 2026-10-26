import assert from "node:assert/strict";
import { test } from "node:test";
import {
	formatTime,
	isWeeklyRecurrence,
	shouldShowAudience,
	weekMarker,
	parseTime,
	positionEntries,
	prepareEntries,
} from "../src/Components/Timetable/timetable.ts";

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

test("settings retains hidden classes at their original slot without duplicating ghosts", () => {
	const overrides = new Map([["lesson", { isHidden: true, startHour: "11:00" }]]);
	const hidden = prepareEntries([lesson()], overrides, true);
	assert.equal(hidden.length, 1);
	assert.equal(hidden[0].hidden, true);
	assert.equal(hidden[0].start, 480);
	assert.equal(prepareEntries([lesson()], overrides).length, 0);
	overrides.set("lesson", { isHidden: true, isTransparent: true });
	const ghost = prepareEntries([lesson()], overrides, true);
	assert.equal(ghost.length, 1);
	assert.equal(ghost[0].hidden, true);
	assert.equal(ghost[0].ghost, true);
	assert.equal(prepareEntries([lesson()], overrides)[0].hidden, false);
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

test("only A/B week markers move to the time footer", () => {
	assert.equal(weekMarker(" A "), "A");
	assert.equal(weekMarker("tydzień B"), "B");
	assert.equal(weekMarker("A, bez 05.10"), undefined);
});

test("hide lecture, own group, aliases, and parent audiences", () => {
	assert.equal(shouldShowAudience("W", "cały rocznik", "GL2"), false);
	assert.equal(shouldShowAudience("wykład", "GL1", "GL2"), false);
	assert.equal(shouldShowAudience("L", "GL2", "GL2"), false);
	assert.equal(shouldShowAudience("L", "GK2", "GL2"), false);
	assert.equal(shouldShowAudience("C", "GC1", "GK2"), false);
	assert.equal(shouldShowAudience("C", "C1 (GL1+GL2)", "GL2"), false);
	assert.equal(shouldShowAudience("C", "GC2", "GL2"), true);
	assert.equal(shouldShowAudience("L", "GL1", "GL2"), true);
	assert.equal(shouldShowAudience("C", "grupa obieralna 1", "GL2"), true);
	assert.equal(shouldShowAudience("L", "GL2"), true);
});

test("default weekly recurrence is recognized as noise, alternating weeks are not", () => {
	assert.equal(isWeeklyRecurrence(undefined), true);
	assert.equal(isWeeklyRecurrence(""), true);
	assert.equal(isWeeklyRecurrence("co tydz."), true);
	assert.equal(isWeeklyRecurrence("co tydzień"), true);
	assert.equal(isWeeklyRecurrence("Co Tydz."), true);
	assert.equal(isWeeklyRecurrence("co 2 tydzień - 15h"), false);
	assert.equal(isWeeklyRecurrence("co 2 tydzień (15h)"), false);
});
