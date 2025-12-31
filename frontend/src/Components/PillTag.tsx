import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { colors } from "../styleTokens.stylex";
import { attrs } from "../utils";

type PillTagProps = {
	isActivable?: boolean;
	isActive?: boolean;
	text: string;
};

const styles = stylex.create({
	metaLabel: {
		borderRadius: "999px",
		paddingBlock: "0.35rem",
		paddingInline: "0.75rem",
		backgroundColor: "#f4f6fb",
		color: colors.secondaryText,

		fontSize: "0.8rem",
	},
	status: {
		fontWeight: 600,
	},
	statusActive: {
		backgroundColor: colors.activeBackground,
		color: colors.activeText,
	},
	statusInactive: {
		backgroundColor: colors.inactiveBackground,
		color: colors.inactiveText,
	},
});

export function PillTag(props: PillTagProps) {
	return (
		<Show when={props.isActivable} fallback={<span {...attrs(stylex.props(styles.metaLabel))}>{props.text}</span>}>
			<span
				{...attrs(
					stylex.props(styles.metaLabel, styles.status, props.isActive ? styles.statusActive : styles.statusInactive)
				)}
			>
				{props.text}
			</span>
		</Show>
	);
}
