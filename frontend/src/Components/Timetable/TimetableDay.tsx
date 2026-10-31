import * as stylex from "@stylexjs/stylex";
import { createMemo, For, Show } from "solid-js";
import { attrs } from "../../utils";
import type { ClassInfo } from "../../gen/pkplaner/v1/pkplaner_pb";
import { positionEntries, type TimetableEntry } from "./timetable";
import { TimetableEvent } from "./TimetableEvent";
import { timetableStyles as styles } from "./timetableStyles";

export function TimetableDay(props: {
	label: string;
	onClassSelect?: (classInfo: ClassInfo) => void;
	entries: TimetableEntry[];
	start: number;
	end: number;
	pixelsPerMinute: number;
	selected: boolean;
	today: boolean;
	now?: number;
}) {
	const entries = createMemo(() => positionEntries(props.entries));
	const intervals = () => Array.from({ length: (props.end - props.start) / 30 }, (_, i) => props.start + i * 30);
	const showNow = () => props.now !== undefined && props.now >= props.start && props.now <= props.end;
	return (
		<section
			{...attrs(stylex.props(styles.day, !props.selected && styles.inactiveDay))}
			aria-label={props.label}
			aria-current={props.today ? "date" : undefined}
		>
			<h3 {...attrs(stylex.props(styles.dayHeader, props.today && styles.todayHeader))}>
				<Show when={props.today}>
					<span {...attrs(stylex.props(styles.todayDot))} aria-hidden="true" />
				</Show>
				{props.label}
			</h3>
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
				<Show when={showNow() ? props.now : undefined}>
					{(now) => (
						<div
							{...attrs(stylex.props(styles.nowLine))}
							style={{ top: `${(now() - props.start) * props.pixelsPerMinute}px` }}
							aria-hidden="true"
						>
							<span {...attrs(stylex.props(styles.nowDot))} />
						</div>
					)}
				</Show>
				<Show when={entries().length === 0}>
					<span {...attrs(stylex.props(styles.srOnly))}>Brak zajęć</span>
				</Show>
				<For each={entries()}>
					{(entry) => (
						<TimetableEvent
							entry={entry}
							onClassSelect={props.onClassSelect}
							start={props.start}
							pixelsPerMinute={props.pixelsPerMinute}
						/>
					)}
				</For>
			</div>
		</section>
	);
}
