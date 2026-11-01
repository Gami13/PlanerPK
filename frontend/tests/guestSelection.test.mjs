import assert from "node:assert/strict";
import { test } from "node:test";
import {
	resolveGuestSelection,
	readGuestSelection,
	saveGuestSelection,
} from "../src/components/Timetable/guestSelection.ts";

test("defaults to the first available timetable and restores valid selections", () => {
	const columns = [{ column: 3 }, { column: 7 }];
	assert.equal(resolveGuestSelection(columns), 3);
	assert.equal(resolveGuestSelection(columns, 7), 7);
	assert.equal(resolveGuestSelection(columns, 99), 3);
	assert.equal(resolveGuestSelection([]), undefined);
});

test("guest selection persists and storage failures do not break browsing", () => {
	const original = Object.getOwnPropertyDescriptor(globalThis, "localStorage");
	try {
		const values = new Map();
		Object.defineProperty(globalThis, "localStorage", {
			configurable: true,
			value: {
				getItem: (key) => values.get(key) ?? null,
				setItem: (key, value) => values.set(key, value),
			},
		});
		assert.equal(readGuestSelection(), undefined);
		saveGuestSelection(7);
		assert.equal(readGuestSelection(), 7);
		Object.defineProperty(globalThis, "localStorage", {
			configurable: true,
			get() {
				throw new Error("blocked");
			},
		});
		assert.equal(readGuestSelection(), undefined);
		assert.doesNotThrow(() => saveGuestSelection(3));
	} finally {
		if (original) Object.defineProperty(globalThis, "localStorage", original);
		else delete globalThis.localStorage;
	}
});
