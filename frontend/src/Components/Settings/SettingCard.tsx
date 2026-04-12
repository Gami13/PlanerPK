import * as stylex from "@stylexjs/stylex";
import type { JSX } from "solid-js";
import { Show } from "solid-js";
import { borderRadius, colors, spacing } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";

type SettingCardProps = {
	title: string;
	description?: string;
	children: JSX.Element;
};
const style = stylex.create({
	setting: {
		padding: spacing.lg,
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.xl,
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		backgroundColor: colors.white,
		display: "flex",
		flexDirection: "column",
		width: "min(100%, 48em)",
	},
	settingHeader: {
		margin: 0,
		fontSize: "1.4em",
		fontWeight: 600,
	},
	settingDescription: {
		margin: 0,
		color: colors.secondaryText,
	},
});

export function SettingCard(props: SettingCardProps) {
	return (
		<article {...attrs(stylex.props(style.setting))}>
			<h2 {...attrs(stylex.props(style.settingHeader))}>{props.title}</h2>
			<Show when={props.description}>
				<p {...attrs(stylex.props(style.settingDescription))}>{props.description}</p>
			</Show>
			{props.children}
		</article>
	);
}
