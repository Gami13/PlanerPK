import assert from "node:assert/strict";
import { test } from "node:test";
import { createModalDialog } from "../src/Components/createModalDialog.ts";

function fixture() {
	let pending = false;
	let closed = 0;
	let notified = 0;
	const element = {
		close: () => closed++,
		getBoundingClientRect: () => ({ left: 10, right: 100, top: 20, bottom: 200 }),
	};
	const dialog = createModalDialog(
		() => pending,
		() => notified++
	);
	dialog.props.ref(element);
	return {
		dialog,
		element,
		closed: () => closed,
		notified: () => notified,
		setPending: (value) => (pending = value),
		click: (clientX, clientY, target = element) => dialog.props.onClick({ target, clientX, clientY }),
	};
}

test("only backdrop clicks outside the dialog bounds dismiss it", () => {
	const f = fixture();
	f.click(50, 100);
	f.click(10, 20);
	f.click(100, 200);
	f.click(0, 0, {}); // A child, not the backdrop.
	assert.equal(f.closed(), 0);
	for (const [x, y] of [
		[9, 100],
		[101, 100],
		[50, 19],
		[50, 201],
	])
		f.click(x, y);
	assert.equal(f.closed(), 4);
});

test("pending operations block backdrop and Escape dismissal, then allow both again", () => {
	const f = fixture();
	f.setPending("lesson-hash"); // AddClassDialog uses a hash; the editor uses a boolean.
	f.click(0, 0);
	const blocked = new Event("cancel", { cancelable: true });
	f.dialog.props.onCancel(blocked);
	assert.equal(f.closed(), 0);
	assert.equal(blocked.defaultPrevented, true);
	f.setPending(false);
	f.click(0, 0);
	const allowed = new Event("cancel", { cancelable: true });
	f.dialog.props.onCancel(allowed);
	assert.equal(f.closed(), 1);
	assert.equal(allowed.defaultPrevented, false);
});

test("explicit close and native close notifications are forwarded", () => {
	const f = fixture();
	f.dialog.close();
	assert.equal(f.closed(), 1);
	f.dialog.props.onClose();
	assert.equal(f.notified(), 1);
});
