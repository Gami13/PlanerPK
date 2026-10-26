import * as stylex from "@stylexjs/stylex";
import { createMemo, createSignal, For, Match, onCleanup, onMount, Show, Switch } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs } from "../../utils";
import type { JSX } from "solid-js";
import { formatTime, prepareEntries, weekdays } from "./timetable";
import { TimetableDay } from "./TimetableDay";
import { TimetableDayNavigation } from "./TimetableDayNavigation";
import { timetableStyles as styles } from "./timetableStyles";

type TimetablePreviewProps = {
	expand?: boolean;
	headerAction?: JSX.Element;
	classes?: ClassInfo[];
	overrides?: Map<string, ClassOverride>;
	loading?: boolean;
	error?: unknown;
	groupLabel?: string;
	noGroup?: boolean;
	showHidden?: boolean;
	onShowHiddenChange?: (show: boolean) => void;
	onClassSelect?: (classInfo: ClassInfo) => void;
};

function currentMinute(): number {
	const now = new Date();
	return now.getHours() * 60 + now.getMinutes();
}

export function TimetablePreview(props: TimetablePreviewProps) {
	const weekday = new Date().getDay();
	const todayIndex = weekday >= 1 && weekday <= 5 ? weekday - 1 : undefined;
	const [selected, setSelected] = createSignal(todayIndex ?? 0);
	const [now, setNow] = createSignal(currentMinute());
	const entries = createMemo(() => prepareEntries(props.classes ?? [], props.overrides, props.showHidden));
	// Begin half an hour before the first labeled hour, leaving room above its centered label.
	const firstHour = createMemo(() => Math.floor(Math.min(8 * 60, ...entries().map((entry) => entry.start)) / 60) * 60);
	const start = createMemo(() => firstHour() - 30);
	const end = createMemo(() => {
		const last = Math.max(...entries().map((entry) => entry.end), 8 * 60);
		// Half an hour of breathing room under the last block, snapped to the grid.
		return Math.ceil((last + 30) / 30) * 30;
	});
	const hours = () => {
		const labels: number[] = [];
		for (let hour = firstHour(); hour < end(); hour += 60) labels.push(hour);
		return labels;
	};
	const pixelsPerMinute = 0.85;

	onMount(() => {
		const timer = setInterval(() => setNow(currentMinute()), 60_000);
		onCleanup(() => clearInterval(timer));
	});

	return (
		<section {...attrs(stylex.props(styles.card))} aria-label="Plan tygodnia" aria-busy={!!props.loading}>
			<header {...attrs(stylex.props(styles.header))}>
				<div {...attrs(stylex.props(styles.headingGroup))}>
					<div {...attrs(stylex.props(styles.titleRow))}>
						<h2 {...attrs(stylex.props(styles.title))}>Plan tygodnia</h2>
						<Show when={props.groupLabel}>
							<span {...attrs(stylex.props(styles.groupLabel))}>· {props.groupLabel}</span>
						</Show>
					</div>
					<Show when={props.onClassSelect}>
						<p {...attrs(stylex.props(styles.subtitle))}>Kliknij zajęcia, aby zmienić ich ustawienia.</p>
					</Show>
				</div>
				{props.headerAction}
				<Show when={props.onShowHiddenChange}>
					<button
						{...attrs(stylex.props(styles.hiddenToggle))}
						type="button"
						aria-pressed={!!props.showHidden}
						onClick={() => props.onShowHiddenChange?.(!props.showHidden)}
					>
						{props.showHidden ? "Ukryj ukryte zajęcia" : "Pokaż ukryte zajęcia"}
					</button>
				</Show>
			</header>
			<Switch
				fallback={
					<>
						<TimetableDayNavigation selected={selected()} onChange={setSelected} />
						<div
							{...attrs(stylex.props(styles.scroll, props.expand && styles.expanded))}
							tabindex="0"
							role="region"
							aria-label="Tygodniowa siatka zajęć"
						>
							<div {...attrs(stylex.props(styles.grid))}>
								<div {...attrs(stylex.props(styles.axis))} aria-hidden="true">
									<div {...attrs(stylex.props(styles.axisHeader))} />
									<div
										{...attrs(stylex.props(styles.timeline))}
										style={{ height: `${(end() - start()) * pixelsPerMinute}px` }}
									>
										<For each={hours()}>
											{(hour) => (
												<span
													{...attrs(stylex.props(styles.hourLabel))}
													style={{ top: `${(hour - start()) * pixelsPerMinute}px` }}
												>
													{formatTime(hour)}
												</span>
											)}
										</For>
									</div>
								</div>
								<For each={weekdays}>
									{(day, index) => (
										<TimetableDay
											label={day.label}
											onClassSelect={props.onClassSelect}
											entries={entries().filter((entry) => entry.classInfo.day === day.key)}
											start={start()}
											end={end()}
											pixelsPerMinute={pixelsPerMinute}
											selected={index() === selected()}
											today={index() === todayIndex}
											now={index() === todayIndex ? now() : undefined}
										/>
									)}
								</For>
							</div>
						</div>
					</>
				}
			>
				<Match when={props.noGroup}>
					<p {...attrs(stylex.props(styles.message))}>Wybierz grupę w ustawieniach, aby zobaczyć plan.</p>
				</Match>
				<Match when={props.loading}>
					<p {...attrs(stylex.props(styles.message))} role="status">
						Ładowanie planu…
					</p>
				</Match>
				<Match when={props.error}>
					<p {...attrs(stylex.props(styles.message))} role="alert">
						Nie udało się załadować planu. Spróbuj ponownie później.
					</p>
				</Match>
				<Match when={entries().length === 0}>
					<p {...attrs(stylex.props(styles.message))}>Brak zajęć do wyświetlenia.</p>
				</Match>
			</Switch>
		</section>
	);
}
