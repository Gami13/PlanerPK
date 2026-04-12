import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { Show, type JSX } from "solid-js";
import { primitiveStyles } from "../styles/primitives";
import { attrs } from "../utils";

const styles = {
	buttonBase: primitiveStyles.buttonBase,
	primaryButton: primitiveStyles.buttonPrimary,
	secondaryButton: primitiveStyles.buttonSecondary,
};

type ButtonProps = {
	children: JSX.Element;
	isPrimary?: boolean;
	onClick?: () => void;
	href?: string;
	isVanillaLink?: boolean;
};

export function Button(props: ButtonProps) {
	const buttonStyles = () =>
		props.isPrimary
			? stylex.props(styles.buttonBase, styles.primaryButton)
			: stylex.props(styles.buttonBase, styles.secondaryButton);
	const buttonAttrs = () => attrs(buttonStyles());

	return (
		<>
			<Show when={props.href && !props.isVanillaLink}>
				<A {...buttonAttrs()} href={props.href!}>
					{props.children}
				</A>
			</Show>
			<Show when={props.href && props.isVanillaLink}>
				<a {...buttonAttrs()} href={props.href}>
					{props.children}
				</a>
			</Show>
			<Show when={!props.href}>
				<button {...buttonAttrs()} onClick={props.onClick}>
					{props.children}
				</button>
			</Show>
		</>
	);
}
