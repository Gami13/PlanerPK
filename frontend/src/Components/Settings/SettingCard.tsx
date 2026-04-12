import * as stylex from "@stylexjs/stylex";
import type { JSX } from "solid-js";
import { Show } from "solid-js";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type SettingCardProps = {
	title: string;
	description?: string;
	children: JSX.Element;
};

export function SettingCard(props: SettingCardProps) {
	return (
		<article {...attrs(stylex.props(settingsStyles.setting))}>
			<h2 {...attrs(stylex.props(settingsStyles.settingHeader))}>{props.title}</h2>
			<Show when={props.description}>
				<p {...attrs(stylex.props(settingsStyles.settingDescription))}>{props.description}</p>
			</Show>
			{props.children}
		</article>
	);
}
