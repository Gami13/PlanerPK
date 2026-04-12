import { createSignal } from "solid-js";

type AsyncStatus = "idle" | "loading" | "success" | "error";

export function useAsyncAction(successMs = 2500, errorMs = 3000) {
	const [status, setStatus] = createSignal<AsyncStatus>("idle");

	const run = async (fn: () => Promise<boolean>) => {
		setStatus("loading");
		const success = await fn();

		if (success) {
			setStatus("success");
			setTimeout(() => setStatus("idle"), successMs);
		} else {
			setStatus("error");
			setTimeout(() => setStatus("idle"), errorMs);
		}
		return success;
	};

	return { status, run };
}
