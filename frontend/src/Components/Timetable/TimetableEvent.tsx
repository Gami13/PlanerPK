import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import type { ClassInfo } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs } from "../../utils";
import { formatTime, isWeeklyRecurrence, weekMarker, type PositionedEntry } from "./timetable";
import { timetableStyles as styles } from "./timetableStyles";
import { defaultTitleTemplate, formatClassTitle } from "./titleTemplate";

export function TimetableEvent(props: {
	entry: PositionedEntry;
	onClassSelect?: (classInfo: ClassInfo) => void;
	start: number;
	pixelsPerMinute: number;
}) {
	const info = () => props.entry.classInfo;
	const duration = () => props.entry.end - props.entry.start;
	// Short blocks only have room for a title and the time; drop the rest
	// instead of letting it clip mid-line.
	const roomy = () => duration() >= 75;
	const title = () => info().summary || formatClassTitle(defaultTitleTemplate, info());
	const heading = () => {
		const suffix = ` - ${info().teacher}`;
		const withoutTeacher = info().teacher && title().endsWith(suffix) ? title().slice(0, -suffix.length) : title();
		const typeSuffix = ` (${info().type})`;
		return info().type && withoutTeacher.endsWith(typeSuffix)
			? withoutTeacher.slice(0, -typeSuffix.length)
			: withoutTeacher;
	};
	const week = () => (props.entry.ghost ? undefined : weekMarker(info().recurrence));
	const showRecurrence = () =>
		!props.entry.ghost && !week() && !!info().recurrence && !isWeeklyRecurrence(info().recurrence);
	const time = () => `${formatTime(props.entry.start)}–${formatTime(props.entry.end)}`;
	const tooltip = () =>
		[title(), info().teacher, info().recurrence, info().location, time()].filter(Boolean).join("\n");

	return (
		<article
			{...attrs(
				stylex.props(
					styles.event,
					props.entry.ghost && styles.ghost,
					props.entry.hidden && styles.hidden,
					!!props.onClassSelect && styles.interactive
				)
			)}
			role={props.onClassSelect ? "button" : undefined}
			tabIndex={props.onClassSelect ? 0 : undefined}
			aria-haspopup={props.onClassSelect ? "dialog" : undefined}
			aria-label={
				props.onClassSelect
					? `${title()}, ${time()}${props.entry.hidden ? ", ukryte" : ""}${props.entry.ghost ? ", transparentne" : ""}. Edytuj zajęcia`
					: undefined
			}
			onClick={() => props.onClassSelect?.(info())}
			onKeyDown={(event) => {
				if (props.onClassSelect && (event.key === "Enter" || event.key === " ")) {
					event.preventDefault();
					props.onClassSelect(info());
				}
			}}
			style={{
				top: `${(props.entry.start - props.start) * props.pixelsPerMinute}px`,
				height: `${duration() * props.pixelsPerMinute - 1}px`,
				left: `calc(${(props.entry.lane / props.entry.lanes) * 100}% + ${1 / props.entry.lanes}%)`,
				width: `calc(${98 / props.entry.lanes}%)`,
			}}
			title={tooltip()}
		>
			<Show when={info().type}>
				<span
					{...attrs(stylex.props(styles.eventType))}
					title={`Typ zajęć: ${info().type}`}
					aria-label={`Typ zajęć: ${info().type}`}
				>
					{info().type}
				</span>
			</Show>
			<div {...attrs(stylex.props(styles.eventContent))}>
				<h4 {...attrs(stylex.props(styles.eventName, !!info().type && styles.eventNameWithType))}>{heading()}</h4>
				<Show when={info().teacher}>
					<span {...attrs(stylex.props(styles.eventTeacher))} title={info().teacher}>
						{info().teacher}
					</span>
				</Show>
				<Show when={props.entry.hidden}>
					<span {...attrs(stylex.props(styles.eventMeta))}>Ukryte{props.entry.ghost ? " · transparentne" : ""}</span>
				</Show>
				<Show when={props.entry.ghost && !props.entry.hidden}>
					<span {...attrs(stylex.props(styles.eventMeta))}>Poprzednia pozycja</span>
				</Show>
				<Show when={showRecurrence() && roomy()}>
					<span
						{...attrs(stylex.props(styles.eventMeta))}
						title={info().recurrence}
						aria-label={`Info: ${info().recurrence}`}
					>
						{info().recurrence}
					</span>
				</Show>
			</div>
			<div {...attrs(stylex.props(styles.eventFooter))}>
				<Show when={info().location}>
					<span {...attrs(stylex.props(styles.eventRoom))} title={info().location}>
						{info().location}
					</span>
				</Show>
				<span {...attrs(stylex.props(styles.eventTime))}>
					<Show when={week()}>
						{(marker) => (
							<span aria-label={`Tydzień ${marker()}`} {...attrs(stylex.props(styles.eventWeek))}>
								{marker()}
							</span>
						)}
					</Show>
					{time()}
				</span>
			</div>
		</article>
	);
}
