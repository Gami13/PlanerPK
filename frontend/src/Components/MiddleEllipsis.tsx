import * as stylex from "@stylexjs/stylex";
import { createEffect, createSignal, onCleanup, onMount } from "solid-js";
import { attrs } from "../utils";
import { middleEllipsis } from "./truncateMiddle";

const styles = stylex.create({
	text: {
		display: "block",
		minWidth: 0,
		whiteSpace: "nowrap",
		overflow: "hidden",
		textOverflow: "ellipsis",
	},
});

export function MiddleEllipsis(props: { text: string }) {
	let element!: HTMLSpanElement;
	let context: CanvasRenderingContext2D | null = null;
	let disposed = false;
	const [displayText, setDisplayText] = createSignal(props.text);
	const update = () => {
		if (!element || !context || disposed) return;
		const font = getComputedStyle(element);
		context.font = `${font.fontWeight} ${font.fontSize} ${font.fontFamily}`;
		setDisplayText(middleEllipsis(props.text, element.clientWidth, (text) => context!.measureText(text).width));
	};
	createEffect(() => {
		setDisplayText(props.text);
		update();
	});
	onMount(() => {
		context = document.createElement("canvas").getContext("2d");
		const observer = new ResizeObserver(update);
		observer.observe(element);
		update();
		void document.fonts.ready.then(update);
		onCleanup(() => {
			disposed = true;
			observer.disconnect();
		});
	});
	return (
		<span ref={element} {...attrs(stylex.props(styles.text))} title={props.text} aria-label={props.text}>
			{displayText()}
		</span>
	);
}
