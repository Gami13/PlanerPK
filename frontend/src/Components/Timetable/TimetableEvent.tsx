import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { attrs } from "../../utils";
import { formatTime, isWeeklyRecurrence, shouldShowAudience, weekMarker, type PositionedEntry } from "./timetable";
import { timetableStyles as styles } from "./timetableStyles";
import { defaultTitleTemplate, formatClassTitle } from "./titleTemplate";

export function TimetableEvent(props: {
	entry: PositionedEntry;
	start: number;
	pixelsPerMinute: number;
	groupLabel?: string;
}) {
	const info = () => props.entry.classInfo;
	const duration = () => props.entry.end - props.entry.start;
	// Short blocks only have room for a title and the time; drop the rest
	// instead of letting it clip mid-line.
	const roomy = () => duration() >= 75;
	const title = () => info().summary || formatClassTitle(defaultTitleTemplate, info());
	const showAudience = () => !props.entry.ghost && shouldShowAudience(info().type, info().audience, props.groupLabel);
	const week = () => (props.entry.ghost ? undefined : weekMarker(info().recurrence));
	const showRecurrence = () =>
		!props.entry.ghost && !week() && !!info().recurrence && !isWeeklyRecurrence(info().recurrence);
	const time = () => `${formatTime(props.entry.start)}–${formatTime(props.entry.end)}`;
	const tooltip = () =>
		[
			title(),
			info().teacher,
			info().audience && `Grupa: ${info().audience}`,
			info().recurrence,
			info().location,
			time(),
		]
			.filter(Boolean)
			.join("\n");

	return (
		<article
			{...attrs(stylex.props(styles.event, props.entry.ghost && styles.ghost))}
			style={{
				top: `${(props.entry.start - props.start) * props.pixelsPerMinute}px`,
				height: `${duration() * props.pixelsPerMinute - 1}px`,
				left: `calc(${(props.entry.lane / props.entry.lanes) * 100}% + ${1 / props.entry.lanes}%)`,
				width: `calc(${98 / props.entry.lanes}%)`,
			}}
			title={tooltip()}
		>
			<div {...attrs(stylex.props(styles.eventContent))}>
				<h4 {...attrs(stylex.props(styles.eventName))}>{title()}</h4>
				<Show when={props.entry.ghost}>
					<span {...attrs(stylex.props(styles.eventMeta))}>Poprzednia pozycja</span>
				</Show>
				<Show when={showAudience() && roomy()}>
					<span
						{...attrs(stylex.props(styles.eventMeta))}
						title={`Grupa: ${info().audience}`}
						aria-label={`Grupa: ${info().audience}`}
					>
						{info().audience}
					</span>
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
