import { onMount, type Accessor } from "solid-js";

export function createModalDialog(pending: Accessor<unknown>, onClose: () => void) {
	let dialog!: HTMLDialogElement;
	onMount(() => dialog.showModal());
	return {
		close: () => dialog.close(),
		props: {
			ref: (element: HTMLDialogElement) => (dialog = element),
			onClose,
			onCancel: (event: Event) => {
				if (pending()) event.preventDefault();
			},
			onClick: (event: MouseEvent) => {
				if (event.target !== dialog || pending()) return;
				// The dialog itself is also the target for clicks inside its padding.
				const bounds = dialog.getBoundingClientRect();
				if (
					event.clientX < bounds.left ||
					event.clientX > bounds.right ||
					event.clientY < bounds.top ||
					event.clientY > bounds.bottom
				)
					dialog.close();
			},
		},
	};
}
