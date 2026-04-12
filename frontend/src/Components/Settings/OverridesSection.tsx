import * as stylex from "@stylexjs/stylex";
import { For, Show } from "solid-js";
import { Button } from "../../Components/Button";
import { DAY_LABELS } from "../../constants";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs } from "../../utils";
import { ClassCard } from "./ClassCard";
import { settingsStyles } from "./settingsStyles";

type OverridesSectionProps = {
	chosenColumn: number;
	classesLoading: boolean;
	classesError: unknown;
	classes: ClassInfo[];
	classesAvailableToAdd: ClassInfo[];
	visibleClasses: ClassInfo[];
	selectedClassHash: string;
	onSelectedClassHashChange: (value: string) => void;
	onAddOverride: () => void;
	overridesByHash: Map<string, ClassOverride> | undefined;
	onOverrideSaved: (updated: ClassOverride) => void;
};

export function OverridesSection(props: OverridesSectionProps) {
	const showOverridesEditor = () => props.chosenColumn > 0 && props.classes.length > 0;

	return (
		<article {...attrs(stylex.props(settingsStyles.setting))}>
			<h2 {...attrs(stylex.props(settingsStyles.settingHeader))}>Nadpisania zajęć</h2>
			<p {...attrs(stylex.props(settingsStyles.settingDescription))}>
				Możesz ukryć zajęcia, przesunąć ich godzinę lub zostawić poprzednią pozycję jako transparentną w kalendarzu.
				Zmiany zostaną zastosowane przy następnej synchronizacji.
			</p>

			<Show when={props.chosenColumn <= 0}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Wybierz kolumnę powyżej, aby zobaczyć listę zajęć.</p>
			</Show>
			<Show when={props.chosenColumn > 0 && props.classesLoading}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Ładowanie zajęć…</p>
			</Show>
			<Show when={props.chosenColumn > 0 && props.classesError}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Błąd ładowania zajęć.</p>
			</Show>
			<Show when={props.chosenColumn > 0 && props.classes.length === 0}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Brak zajęć dla wybranej kolumny.</p>
			</Show>

			<Show when={showOverridesEditor()}>
				<div {...attrs(stylex.props(settingsStyles.addOverrideRow))}>
					<select
						id="override-class-select"
						value={props.selectedClassHash}
						onChange={(e) => props.onSelectedClassHashChange(e.currentTarget.value)}
						{...attrs(stylex.props(settingsStyles.addOverrideSelect))}
					>
						<option value="">Wybierz zajęcia do nadpisania…</option>
						<For each={props.classesAvailableToAdd}>
							{(cls) => (
								<option value={cls.hash}>
									{cls.name}
									<Show when={cls.type}> ({cls.type})</Show>
									{" - "}
									{DAY_LABELS[cls.day] ?? cls.day} {cls.startTime}-{cls.endTime} -- {cls.teacher} @ {cls.location}
								</option>
							)}
						</For>
					</select>
					<Button isPrimary={false} onClick={props.onAddOverride}>
						Dodaj nadpisanie
					</Button>
				</div>

				<div {...attrs(stylex.props(settingsStyles.overrideGrid))}>
					<For each={props.visibleClasses}>
						{(cls) => (
							<ClassCard
								classInfo={cls}
								override={props.overridesByHash?.get(cls.hash)}
								onSaved={props.onOverrideSaved}
							/>
						)}
					</For>
					<Show when={props.visibleClasses.length === 0}>
						<p {...attrs(stylex.props(settingsStyles.emptyNote))}>
							Brak aktywnych nadpisań. Wybierz zajęcia z listy powyżej.
						</p>
					</Show>
				</div>
			</Show>
		</article>
	);
}
