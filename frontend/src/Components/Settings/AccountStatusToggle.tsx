import * as stylex from "@stylexjs/stylex";
import { createSignal } from "solid-js";
import { UIText } from "../../constants";
import { borderRadius, colors } from "../../styles/tokens.stylex";
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
		width: "48px",
		height: "28px",
		padding: "3px",
		borderWidth: 0,
		borderRadius: borderRadius.full,
		backgroundColor: colors.secondaryText,
		cursor: "pointer",
		transition: "background-color 0.15s ease",
		outline: { default: "none", ":focus-visible": "2px solid #039be5" },
		outlineOffset: "3px",
	},
	active: {
		backgroundColor: colors.primary,
	},
	disabled: {
		cursor: "wait",
		opacity: 0.6,
	},
	thumb: {
		width: "22px",
		height: "22px",
		borderRadius: borderRadius.full,
		backgroundColor: colors.white,
		transform: "translateX(0)",
		transition: "transform 0.15s ease",
	},
	thumbActive: {
		transform: "translateX(20px)",
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
