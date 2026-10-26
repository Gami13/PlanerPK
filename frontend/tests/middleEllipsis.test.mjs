import assert from "node:assert/strict";
import { test } from "node:test";
import { middleEllipsis } from "../src/Components/truncateMiddle.ts";

const measure = (text) => Array.from(text).length;

test("short titles are unchanged", () => {
	assert.equal(middleEllipsis("Algebra", 10, measure), "Algebra");
});

test("long titles retain both ends within the available width", () => {
	assert.equal(middleEllipsis("Matematyka dyskretna (L)", 11, measure), "Matem…a (L)");
});

test("handles very narrow columns and Unicode", () => {
	assert.equal(middleEllipsis("Algebra", 1, measure), "…");
	assert.equal(middleEllipsis("Algebra", 0, measure), "");
	assert.equal(middleEllipsis("😀abcdef😀", 5, measure), "😀a…f😀");
});
