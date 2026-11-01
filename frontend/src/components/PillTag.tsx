import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { primitiveStyles } from "../styles/primitives";
import { attrs } from "../utils";

type PillTagProps = {
	isActivable?: boolean;
	isActive?: boolean;
	text: string;
};

const styles = {
	metaLabel: primitiveStyles.pillBase,
	status: primitiveStyles.pillStatusText,
	statusActive: primitiveStyles.pillStatusActive,
	statusInactive: primitiveStyles.pillStatusInactive,
};

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
