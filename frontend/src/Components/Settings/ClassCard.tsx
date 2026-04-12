import * as stylex from "@stylexjs/stylex";
import { Show, createSignal } from "solid-js";
import { Button } from "../../Components/Button";
import { DAY_LABELS, UI_MESSAGES } from "../../constants";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { setClassOverride } from "../../services/pkplaner";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type ClassCardProps = {
	classInfo: ClassInfo;
	override: ClassOverride | undefined;
	onSaved: (updated: ClassOverride) => void;
};

type CardState = {
	isHidden: boolean;
	isTransparent: boolean;
	startHour: string;
	status: "idle" | "saving" | "ok" | "error";
};

export function ClassCard(props: ClassCardProps) {
	const [cardState, setCardState] = createSignal<CardState>({
		isHidden: props.override?.isHidden ?? false,
		isTransparent: props.override?.isTransparent ?? false,
		startHour: props.override?.startHour ?? "",
		status: "idle",
	});

	const hasActiveOverride = () => {
		const state = cardState();
		return state.isHidden || state.isTransparent || state.startHour !== "";
	};

	async function save() {
		setCardState((prev) => ({ ...prev, status: "saving" }));

		const state = cardState();
		const updatedOverride = {
			classHash: props.classInfo.hash,
			startHour: state.startHour,
			isHidden: state.isHidden,
			isTransparent: state.isTransparent,
		} as ClassOverride;

		const success = await setClassOverride(updatedOverride);
		if (success) {
			setCardState((prev) => ({ ...prev, status: "ok" }));
			props.onSaved(updatedOverride);
			setTimeout(() => setCardState((prev) => ({ ...prev, status: "idle" })), 2500);
		} else {
			setCardState((prev) => ({ ...prev, status: "error" }));
			setTimeout(() => setCardState((prev) => ({ ...prev, status: "idle" })), 3000);
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
							checked={cardState().isHidden}
							onChange={(e) =>
								setCardState((prev) => ({
									...prev,
									isHidden: e.currentTarget.checked,
								}))
							}
						/>
						{UI_MESSAGES.HIDE_CLASS}
					</label>
					<label
						{...attrs(stylex.props(settingsStyles.checkboxLabel))}
						id={`transparent-label-${props.classInfo.hash}`}
					>
						<input
							id={`transparent-${props.classInfo.hash}`}
							type="checkbox"
							checked={cardState().isTransparent}
							onChange={(e) =>
								setCardState((prev) => ({
									...prev,
									isTransparent: e.currentTarget.checked,
								}))
							}
						/>
						{UI_MESSAGES.SHOW_TRANSPARENT}
					</label>
				</div>

				<Show when={!cardState().isHidden}>
					<div {...attrs(stylex.props(settingsStyles.timeGroup))}>
						<label {...attrs(stylex.props(settingsStyles.timeLabel))} for={`start-hour-${props.classInfo.hash}`}>
							{UI_MESSAGES.SHIFT_TIME}
						</label>
						<input
							id={`start-hour-${props.classInfo.hash}`}
							type="time"
							value={cardState().startHour}
							onInput={(e) =>
								setCardState((prev) => ({
									...prev,
									startHour: e.currentTarget.value,
								}))
							}
							{...attrs(stylex.props(settingsStyles.timeInput))}
						/>
					</div>
				</Show>
			</div>

			<div {...attrs(stylex.props(settingsStyles.saveRow))}>
				<Button isPrimary={true} onClick={save}>
					{cardState().status === "saving" ? UI_MESSAGES.SAVING : UI_MESSAGES.SAVE}
				</Button>
				<Show when={cardState().status === "ok"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackSuccess))}>{UI_MESSAGES.SAVED}</span>
				</Show>
				<Show when={cardState().status === "error"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackError))}>{UI_MESSAGES.SAVE_ERROR}</span>
				</Show>
			</div>
		</div>
	);
}
