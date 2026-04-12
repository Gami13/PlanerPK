import * as stylex from "@stylexjs/stylex";
import { Show, createSignal } from "solid-js";
import { Button } from "../../Components/Button";
import { DAY_LABELS } from "../../constants";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { setClassOverride } from "../../services/pkplaner";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type ClassCardProps = {
	classInfo: ClassInfo;
	override: ClassOverride | undefined;
	onSaved: (updated: ClassOverride) => void;
};

export function ClassCard(props: ClassCardProps) {
	const initial = () => props.override;

	const [isHidden, setIsHidden] = createSignal(initial()?.isHidden ?? false);
	const [isTransparent, setIsTransparent] = createSignal(initial()?.isTransparent ?? false);
	const [startHour, setStartHour] = createSignal(initial()?.startHour ?? "");
	const [status, setStatus] = createSignal<"idle" | "saving" | "ok" | "error">("idle");

	const hasActiveOverride = () => isHidden() || isTransparent() || startHour() !== "";

	async function save() {
		setStatus("saving");
		const updatedOverride = {
			classHash: props.classInfo.hash,
			startHour: startHour(),
			isHidden: isHidden(),
			isTransparent: isTransparent(),
		} as ClassOverride;

		const success = await setClassOverride(updatedOverride);
		if (success) {
			setStatus("ok");
			props.onSaved(updatedOverride);
			setTimeout(() => setStatus("idle"), 2500);
		} else {
			setStatus("error");
			setTimeout(() => setStatus("idle"), 3000);
		}
	}

	const cardStyle = () =>
		hasActiveOverride()
			? stylex.props(settingsStyles.classCard, settingsStyles.classCardActive)
			: stylex.props(settingsStyles.classCard);

	return (
		<div {...attrs(cardStyle())}>
			<div {...attrs(stylex.props(settingsStyles.classCardRow))}>
				<span {...attrs(stylex.props(settingsStyles.classTitle))}>
					{props.classInfo.name}
					<Show when={props.classInfo.type}> ({props.classInfo.type})</Show>
				</span>
				<span {...attrs(stylex.props(settingsStyles.classMeta))}>
					{DAY_LABELS[props.classInfo.day] ?? props.classInfo.day} &nbsp;·&nbsp;
					{props.classInfo.startTime}–{props.classInfo.endTime}
					<Show when={props.classInfo.teacher}> &nbsp;·&nbsp; {props.classInfo.teacher}</Show>
					<Show when={props.classInfo.location}> &nbsp;·&nbsp; {props.classInfo.location}</Show>
				</span>
			</div>

			<div {...attrs(stylex.props(settingsStyles.overrideControls))}>
				<div {...attrs(stylex.props(settingsStyles.checkboxGroup))}>
					<label {...attrs(stylex.props(settingsStyles.checkboxLabel))} id={`hidden-label-${props.classInfo.hash}`}>
						<input
							id={`hidden-${props.classInfo.hash}`}
							type="checkbox"
							checked={isHidden()}
							onChange={(e) => setIsHidden(e.currentTarget.checked)}
						/>
						Ukryj zajęcia
					</label>
					<label
						{...attrs(stylex.props(settingsStyles.checkboxLabel))}
						id={`transparent-label-${props.classInfo.hash}`}
					>
						<input
							id={`transparent-${props.classInfo.hash}`}
							type="checkbox"
							checked={isTransparent()}
							onChange={(e) => setIsTransparent(e.currentTarget.checked)}
						/>
						Pokaż poprzednią pozycję jako transparentną
					</label>
				</div>

				<Show when={!isHidden()}>
					<div {...attrs(stylex.props(settingsStyles.timeGroup))}>
						<label {...attrs(stylex.props(settingsStyles.timeLabel))} for={`start-hour-${props.classInfo.hash}`}>
							Przesuń na godzinę
						</label>
						<input
							id={`start-hour-${props.classInfo.hash}`}
							type="time"
							value={startHour()}
							onInput={(e) => setStartHour(e.currentTarget.value)}
							{...attrs(stylex.props(settingsStyles.timeInput))}
						/>
					</div>
				</Show>
			</div>

			<div {...attrs(stylex.props(settingsStyles.saveRow))}>
				<Button isPrimary={true} onClick={save}>
					{status() === "saving" ? "Zapisywanie…" : "Zapisz"}
				</Button>
				<Show when={status() === "ok"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackSuccess))}>✓ Zapisano</span>
				</Show>
				<Show when={status() === "error"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackError))}>✗ Błąd zapisu</span>
				</Show>
			</div>
		</div>
	);
}
