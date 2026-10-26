import assert from "node:assert/strict";
import { test } from "node:test";
import { defaultTitleTemplate, formatClassTitle } from "../src/Components/Timetable/titleTemplate.ts";

const lesson = { name: "Systemy operacyjne", type: "P", teacher: "Wojtas Jan", audience: "GL2", recurrence: "A" };

test("default template omits audience", () => {
	assert.equal(formatClassTitle(defaultTitleTemplate, lesson), "Systemy operacyjne (P) - Wojtas Jan");
});

test("audience is still available in custom templates", () => {
	assert.equal(formatClassTitle("[{ClassAudience}]", lesson), "[GL2]");
});

test("empty fields keep all punctuation and whitespace", () => {
	assert.equal(formatClassTitle(defaultTitleTemplate, { name: "Algebra" }), "Algebra () - ");
});

test("unknown placeholders are ignored and repeated placeholders are replaced", () => {
	assert.equal(
		formatClassTitle("{Unknown} {ClassTitle}/{ClassTitle} {constructor}", lesson),
		"{Unknown} Systemy operacyjne/Systemy operacyjne {constructor}"
	);
});

test("replacement values are literal and not recursively expanded", () => {
	assert.equal(formatClassTitle("{ClassTitle}", { ...lesson, name: "$& {ClassTeacher}" }), "$& {ClassTeacher}");
});

test("supports empty templates and literal text", () => {
	assert.equal(formatClassTitle("", lesson), "");
	assert.equal(formatClassTitle("My class", lesson), "My class");
});
