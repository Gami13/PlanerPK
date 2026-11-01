import * as stylex from "@stylexjs/stylex";
import { createSignal } from "solid-js";
import { UIText } from "../../constants";
import { borderRadius } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";

type AccountStatusToggleProps = {
	isActive: boolean | undefined;
	onToggle: () => Promise<void>;
};

const styles = stylex.create({
	switch: {
		display: "inline-flex",
		alignItems: "center",
		flexShrink: 0,
		width: "40px",
		height: "24px",
		padding: "3px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: "#d5dde6",
		borderRadius: borderRadius.full,
		backgroundColor: "#f0f3f7",
		cursor: "pointer",
		transition: "background-color 0.15s ease, border-color 0.15s ease",
		outline: { default: "none", ":focus-visible": "2px solid #4b99d2" },
		outlineOffset: "3px",
	},
	active: {
		backgroundColor: "#e4eefb",
		borderColor: "#b7d1ef",
	},
	disabled: {
		cursor: "wait",
		opacity: 0.6,
	},
	thumb: {
		width: "16px",
		height: "16px",
		flexShrink: 0,
		borderRadius: borderRadius.full,
		backgroundColor: "#8794a5",
		transform: "translateX(0)",
		transition: "transform 0.15s ease, background-color 0.15s ease",
	},
	thumbActive: {
		transform: "translateX(16px)",
		backgroundColor: "#4b99d2",
	},
});

export function AccountStatusToggle(props: AccountStatusToggleProps) {
	const [pending, setPending] = createSignal(false);
	const disabled = () => pending() || props.isActive === undefined;

	async function toggle() {
		if (disabled()) return;
		setPending(true);
		try {
			await props.onToggle();
		} finally {
			setPending(false);
		}
	}

	return (
		<button
			{...attrs(stylex.props(styles.switch, props.isActive && styles.active, disabled() && styles.disabled))}
			type="button"
			role="switch"
			aria-checked={!!props.isActive}
			aria-label={UIText.settings.account.title}
			aria-busy={pending()}
			disabled={disabled()}
			onClick={toggle}
		>
			<span aria-hidden="true" {...attrs(stylex.props(styles.thumb, props.isActive && styles.thumbActive))} />
		</button>
	);
}
