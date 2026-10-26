import * as stylex from "@stylexjs/stylex";
import { createMemo, For, Show } from "solid-js";
import { attrs } from "../../utils";
import { positionEntries, type TimetableEntry } from "./timetable";
import { TimetableEvent } from "./TimetableEvent";
import { timetableStyles as styles } from "./timetableStyles";

export function TimetableDay(props: {
	label: string;
	entries: TimetableEntry[];
	start: number;
	end: number;
	pixelsPerMinute: number;
	selected: boolean;
}) {
	const entries = createMemo(() => positionEntries(props.entries));
	const intervals = () => Array.from({ length: (props.end - props.start) / 30 }, (_, i) => props.start + i * 30);
	return (
		<section {...attrs(stylex.props(styles.day, !props.selected && styles.inactiveDay))} aria-label={props.label}>
			<h3 {...attrs(stylex.props(styles.dayHeader))}>{props.label}</h3>
			<div
				{...attrs(stylex.props(styles.timeline))}
				style={{ height: `${(props.end - props.start) * props.pixelsPerMinute}px` }}
			>
				<For each={intervals()}>
					{(minute) => (
						<div
							aria-hidden="true"
							{...attrs(stylex.props(styles.hourLine, minute % 60 !== 0 && styles.halfHourLine))}
							style={{ top: `${(minute - props.start) * props.pixelsPerMinute}px` }}
						/>
					)}
				</For>
				<Show when={entries().length === 0}>
					<span {...attrs(stylex.props(styles.emptyDay))}>Brak zajęć</span>
				</Show>
				<For each={entries()}>
					{(entry) => <TimetableEvent entry={entry} start={props.start} pixelsPerMinute={props.pixelsPerMinute} />}
				</For>
			</div>
		</section>
	);
}
