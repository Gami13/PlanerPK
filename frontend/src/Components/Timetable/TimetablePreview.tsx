import * as stylex from "@stylexjs/stylex";
import { createMemo, createSignal, For, Match, Switch } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs } from "../../utils";
import { formatTime, prepareEntries, weekdays } from "./timetable";
import { TimetableDay } from "./TimetableDay";
import { TimetableDayNavigation } from "./TimetableDayNavigation";
import { timetableStyles as styles } from "./timetableStyles";

type TimetablePreviewProps = {
	classes?: ClassInfo[];
	overrides?: Map<string, ClassOverride>;
	loading?: boolean;
	error?: unknown;
	groupLabel?: string;
	noGroup?: boolean;
};

export function TimetablePreview(props: TimetablePreviewProps) {
	const today = new Date().getDay();
	const [selected, setSelected] = createSignal(today >= 1 && today <= 5 ? today - 1 : 0);
	const entries = createMemo(() => prepareEntries(props.classes ?? [], props.overrides));
	// Begin half an hour before the first labeled hour, leaving room above its centered label.
	const firstHour = createMemo(() => Math.floor(Math.min(8 * 60, ...entries().map((entry) => entry.start)) / 60) * 60);
	const start = createMemo(() => firstHour() - 30);
	const end = createMemo(() => Math.ceil(Math.max(18 * 60, ...entries().map((entry) => entry.end)) / 60) * 60);
	const hours = () => Array.from({ length: (end() - firstHour()) / 60 }, (_, i) => firstHour() + i * 60);
	const pixelsPerMinute = 1;

	return (
		<section {...attrs(stylex.props(styles.card))} aria-label="Plan tygodnia" aria-busy={!!props.loading}>
			<header {...attrs(stylex.props(styles.header))}>
				<h2 {...attrs(stylex.props(styles.title))}>Plan tygodnia</h2>
				<p {...attrs(stylex.props(styles.subtitle))}>
					{props.groupLabel ? `${props.groupLabel} · ` : ""}Podgląd z zapisanymi zmianami
				</p>
			</header>
			<Switch
				fallback={
					<>
						<TimetableDayNavigation selected={selected()} onChange={setSelected} />
						<div
							{...attrs(stylex.props(styles.scroll))}
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
											entries={entries().filter((entry) => entry.classInfo.day === day.key)}
											start={start()}
											end={end()}
											pixelsPerMinute={pixelsPerMinute}
											selected={index() === selected()}
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
