import * as stylex from "@stylexjs/stylex";
import { For, Show, useContext } from "solid-js";
import { Button } from "../../Components/Button";
import { DayLabels, UIText } from "../../constants";
import { attrs } from "../../utils";
import { ClassCard } from "./ClassCard";
import { OverridesContext } from "./OverridesContext";
import { SettingCard } from "./SettingCard";
import { settingsStyles } from "./settingsStyles";

type OverridesSectionProps = {
	chosenColumn: number;
	classesLoading: boolean;
	classesError: unknown;
};

export function OverridesSection(props: OverridesSectionProps) {
	const context = useContext(OverridesContext);
	if (!context) throw new Error("OverridesSection must be used within OverridesProvider");

	const showOverridesEditor = () => props.chosenColumn > 0 && (context.classes() ?? []).length > 0;

	return (
		<SettingCard title={UIText.settings.classOverrides.title} description={UIText.settings.classOverrides.description}>
			<Show when={props.chosenColumn <= 0}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Wybierz kolumnę powyżej, aby zobaczyć listę zajęć.</p>
			</Show>
			<Show when={props.chosenColumn > 0 && props.classesLoading}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>{UIText.settings.classOverrides.loadingClasses}</p>
			</Show>
			<Show when={props.chosenColumn > 0 && props.classesError}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>{UIText.settings.classOverrides.errorLoadingClasses}</p>
			</Show>
			<Show when={props.chosenColumn > 0 && (context.classes() ?? []).length === 0}>
				<p {...attrs(stylex.props(settingsStyles.emptyNote))}>Brak zajęć dla wybranej kolumny.</p>
			</Show>

			<Show when={showOverridesEditor()}>
				<div {...attrs(stylex.props(settingsStyles.addOverrideRow))}>
					<select
						id="override-class-select"
						value={context.selected()}
						onChange={(e) => context.setSelected(e.currentTarget.value)}
						{...attrs(stylex.props(settingsStyles.columnSelect, settingsStyles.addOverrideSelect))}
					>
						<option value="">Wybierz zajęcia do nadpisania…</option>
						<For each={context.availableClasses()}>
							{(cls) => (
								<option value={cls.hash}>
									{cls.name}
									<Show when={cls.type}> ({cls.type})</Show>
									{" - "}
									{DayLabels[cls.day] ?? cls.day} {cls.startTime}-{cls.endTime} -- {cls.teacher} @ {cls.location}
								</option>
							)}
						</For>
					</select>
					<Button isPrimary={false} onClick={context.addSelected}>
						Dodaj nadpisanie
					</Button>
				</div>

				<div {...attrs(stylex.props(settingsStyles.overrideGrid))}>
					<For each={context.visibleClasses()}>
						{(cls) => (
							<ClassCard
								classInfo={cls}
								override={context.overrides()?.get(cls.hash)}
								onSaved={context.handleOverrideSaved}
							/>
						)}
					</For>
					<Show when={context.visibleClasses().length === 0}>
						<p {...attrs(stylex.props(settingsStyles.emptyNote))}>
							Brak aktywnych nadpisań. Wybierz zajęcia z listy powyżej.
						</p>
					</Show>
				</div>
			</Show>
		</SettingCard>
	);
}
