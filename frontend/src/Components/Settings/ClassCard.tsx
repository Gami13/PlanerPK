import * as stylex from "@stylexjs/stylex";
import { Show, createSignal } from "solid-js";
import { Button } from "../../Components/Button";
import { DayLabels, UIText } from "../../constants";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { setClassOverride } from "../../services/pkplaner";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";
import { useAsyncAction } from "./useAsyncAction";

type ClassCardProps = {
	classInfo: ClassInfo;
	override: ClassOverride | undefined;
	onSaved: (updated: ClassOverride) => void;
};

type CardOverride = {
	isHidden: boolean;
	isTransparent: boolean;
	startHour: string;
};

export function ClassCard(props: ClassCardProps) {
	const [localOverride, setLocalOverride] = createSignal<CardOverride>({
		isHidden: props.override?.isHidden ?? false,
		isTransparent: props.override?.isTransparent ?? false,
		startHour: props.override?.startHour ?? "",
	});

	const { status, run } = useAsyncAction();
	const hiddenId = `hidden-${props.classInfo.hash}`;
	const hiddenLabelId = `hidden-label-${props.classInfo.hash}`;
	const transparentId = `transparent-${props.classInfo.hash}`;
	const transparentLabelId = `transparent-label-${props.classInfo.hash}`;
	const startHourId = `start-hour-${props.classInfo.hash}`;
	const classTitleId = `class-title-${props.classInfo.hash}`;

	const updateOverride = (patch: Partial<CardOverride>) => {
		setLocalOverride((prev) => ({ ...prev, ...patch }));
	};

	const hasActiveOverride = () => {
		const o = localOverride();
		return o.isHidden || o.isTransparent || o.startHour !== "";
	};

	async function save() {
		const o = localOverride();
		const updatedOverride = {
			classHash: props.classInfo.hash,
			startHour: o.startHour,
			isHidden: o.isHidden,
			isTransparent: o.isTransparent,
		} as ClassOverride;

		const success = await run(() => setClassOverride(updatedOverride));
		if (success) {
			props.onSaved(updatedOverride);
		}
	}

	const cardStyle = () =>
		hasActiveOverride()
			? stylex.props(settingsStyles.classCard, settingsStyles.classCardActive)
			: stylex.props(settingsStyles.classCard);

	return (
		<article {...attrs(cardStyle())} aria-labelledby={classTitleId}>
			<header {...attrs(stylex.props(settingsStyles.classCardRow))}>
				<h3 {...attrs(stylex.props(settingsStyles.classTitle))} id={classTitleId}>
					{props.classInfo.name}
					<Show when={props.classInfo.type}> ({props.classInfo.type})</Show>
				</h3>
				<p {...attrs(stylex.props(settingsStyles.classMeta))}>
					{DayLabels[props.classInfo.day] ?? props.classInfo.day} &nbsp;·&nbsp;
					<time dateTime={props.classInfo.startTime}>{props.classInfo.startTime}</time>–
					<time dateTime={props.classInfo.endTime}>{props.classInfo.endTime}</time>
					<Show when={props.classInfo.teacher}> &nbsp;·&nbsp; {props.classInfo.teacher}</Show>
					<Show when={props.classInfo.audience}> &nbsp;·&nbsp; {props.classInfo.audience}</Show>
					<Show when={props.classInfo.location}> &nbsp;·&nbsp; {props.classInfo.location}</Show>
				</p>
			</header>

			<section {...attrs(stylex.props(settingsStyles.overrideControls))}>
				<section {...attrs(stylex.props(settingsStyles.checkboxGroup))}>
					<label {...attrs(stylex.props(settingsStyles.checkboxLabel))} id={hiddenLabelId}>
						<input
							id={hiddenId}
							type="checkbox"
							checked={localOverride().isHidden}
							onChange={(e) => updateOverride({ isHidden: e.currentTarget.checked })}
						/>
						{UIText.settings.classOverrides.hideClass}
					</label>
					<label {...attrs(stylex.props(settingsStyles.checkboxLabel))} id={transparentLabelId}>
						<input
							id={transparentId}
							type="checkbox"
							checked={localOverride().isTransparent}
							onChange={(e) => updateOverride({ isTransparent: e.currentTarget.checked })}
						/>
						{UIText.settings.classOverrides.showTransparent}
					</label>
				</section>

				<Show when={!localOverride().isHidden}>
					<section {...attrs(stylex.props(settingsStyles.timeGroup))}>
						<label {...attrs(stylex.props(settingsStyles.timeLabel))} for={startHourId}>
							{UIText.settings.classOverrides.shiftTime}
						</label>
						<input
							id={startHourId}
							type="time"
							value={localOverride().startHour}
							onInput={(e) => updateOverride({ startHour: e.currentTarget.value })}
							{...attrs(stylex.props(settingsStyles.timeInput))}
						/>
					</section>
				</Show>
			</section>

			<footer {...attrs(stylex.props(settingsStyles.saveRow))}>
				<Button isPrimary={true} onClick={save}>
					{status() === "loading" ? UIText.settings.classOverrides.saving : UIText.settings.classOverrides.save}
				</Button>
				<Show when={status() === "success"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackSuccess))}>{UIText.settings.classOverrides.saved}</span>
				</Show>
				<Show when={status() === "error"}>
					<span {...attrs(stylex.props(settingsStyles.feedbackError))}>{UIText.settings.classOverrides.saveError}</span>
				</Show>
			</footer>
		</article>
	);
}
