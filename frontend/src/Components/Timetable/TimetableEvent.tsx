import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { MiddleEllipsis } from "../MiddleEllipsis";
import { attrs } from "../../utils";
import { formatTime, type PositionedEntry } from "./timetable";
import { timetableStyles as styles } from "./timetableStyles";

export function TimetableEvent(props: { entry: PositionedEntry; start: number; pixelsPerMinute: number }) {
	const info = () => props.entry.classInfo;
	const title = () => info().summary || `${info().name}${info().type ? ` (${info().type})` : ""}`;
	return (
		<article
			{...attrs(stylex.props(styles.event, props.entry.ghost && styles.ghost))}
			style={{
				top: `${(props.entry.start - props.start) * props.pixelsPerMinute}px`,
				height: `${(props.entry.end - props.entry.start) * props.pixelsPerMinute - 2}px`,
				left: `calc(${(props.entry.lane / props.entry.lanes) * 100}% + 2px)`,
				width: `calc(${100 / props.entry.lanes}% - 4px)`,
			}}
		>
			<div {...attrs(stylex.props(styles.eventContent))}>
				<h4 {...attrs(stylex.props(styles.eventName))}>
					<MiddleEllipsis text={title()} />
				</h4>
				<Show when={!info().summary && info().teacher}>
					<span {...attrs(stylex.props(styles.eventMeta))} title={info().teacher}>
						{info().teacher}
					</span>
				</Show>
				<Show when={props.entry.ghost}>
					<span>Poprzednia pozycja · transparentna</span>
				</Show>
			</div>
			<div {...attrs(stylex.props(styles.eventFooter))}>
				<Show when={info().location}>
					<span {...attrs(stylex.props(styles.eventRoom))} title={info().location}>
						{info().location}
					</span>
				</Show>
				<span {...attrs(stylex.props(styles.eventTime))}>
					{formatTime(props.entry.start)}–{formatTime(props.entry.end)}
				</span>
			</div>
		</article>
	);
}
