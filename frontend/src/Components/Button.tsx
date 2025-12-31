import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { colors } from "../styleTokens.stylex";
import { attrs } from "../utils";

export const styles = stylex.create({
	buttonBase: {
		borderRadius: "9999px",
		borderStyle: "none",
		borderWidth: 0,
		paddingBlock: "0.45em",
		paddingInline: "1.3em",

		textDecoration: "none",
		transition: "transform 0.15s ease, box-shadow 0.15s ease",
		alignItems: "center",
		cursor: "pointer",
		display: "inline-flex",
		fontSize: "1em",
		fontWeight: 500,
		justifyContent: "center",
		textAlign: "center",
	},
	primaryButton: {
		backgroundColor: colors.primary,
		color: "#fff",
	},
	secondaryButton: {
		borderColor: colors.defaultBorder,
		borderStyle: "solid",
		borderWidth: "1px",
		backgroundColor: "#fff",
		color: colors.primaryText,
	},
});

type ButtonBaseProps = {
	isPrimary?: boolean;
	children: Element | string;
};

type NavigationButtonProps = ButtonBaseProps & {
	isNavigation: true;
	href: string | undefined;
	onClick?: never;
	isVanillaNavigation?: boolean;
};

type ActionButtonProps = ButtonBaseProps & {
	isNavigation?: false;
	onClick?: () => void;
	href?: never;
};

type ButtonProps = NavigationButtonProps | ActionButtonProps;

export function Button(props: ButtonProps) {
	const buttonStyles = () =>
		props.isPrimary
			? stylex.props(styles.buttonBase, styles.primaryButton)
			: stylex.props(styles.buttonBase, styles.secondaryButton);

	return (
		<Show
			when={props.isNavigation}
			fallback={
				<button {...attrs(buttonStyles())} onClick={props.onClick}>
					{props.children}
				</button>
			}
		>
			<Show
				when={props.isNavigation && props.isVanillaNavigation}
				fallback={
					<A {...attrs(buttonStyles())} href={props.href!}>
						{props.children}
					</A>
				}
			>
				<a {...attrs(buttonStyles())} href={props.href!}>
					{props.children}
				</a>
			</Show>
		</Show>
	);
}
