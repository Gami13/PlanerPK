import * as stylex from "@stylexjs/stylex";
import type { JSX } from "solid-js";
import { Show } from "solid-js";
import { colors, spacing } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";

type SettingCardProps = {
	title: string;
	description?: string;
	children?: JSX.Element;
	headerAction?: JSX.Element;
};
const style = stylex.create({
	setting: {
		padding: spacing.md,
		borderColor: colors.defaultBorder,
		borderRadius: "6px",
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		backgroundColor: colors.white,
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	headerRow: {
		display: "flex",
		alignItems: "center",
		justifyContent: "space-between",
		gap: spacing.md,
	},
	settingHeader: {
		margin: 0,
		fontSize: "1rem",
		fontWeight: 600,
	},
	settingDescription: {
		fontSize: "0.85rem",
		margin: 0,
		color: colors.secondaryText,
	},
});

export function SettingCard(props: SettingCardProps) {
	return (
		<article {...attrs(stylex.props(style.setting))}>
			<div {...attrs(stylex.props(style.headerRow))}>
				<h2 {...attrs(stylex.props(style.settingHeader))}>{props.title}</h2>
				{props.headerAction}
			</div>
			<Show when={props.description}>
				<p {...attrs(stylex.props(style.settingDescription))}>{props.description}</p>
			</Show>
			{props.children}
		</article>
	);
}
