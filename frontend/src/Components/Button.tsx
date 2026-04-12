import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { Match, Switch, type JSX } from "solid-js";
import { primitiveStyles } from "../styles/primitives";
import { attrs } from "../utils";

const styles = {
	buttonBase: primitiveStyles.buttonBase,
	primaryButton: primitiveStyles.buttonPrimary,
	secondaryButton: primitiveStyles.buttonSecondary,
};

type ButtonBaseProps = {
	isPrimary?: boolean;
	children: JSX.Element;
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
	const buttonAttrs = () => attrs(buttonStyles());

	return (
		<Switch>
			<Match when={props.isNavigation && props.isVanillaNavigation}>
				<a {...buttonAttrs()} href={props.href ?? "#"}>
					{props.children}
				</a>
			</Match>
			<Match when={props.isNavigation}>
				<A {...buttonAttrs()} href={props.href ?? "#"}>
					{props.children}
				</A>
			</Match>
			<Match when={!props.isNavigation}>
				<button {...buttonAttrs()} onClick={() => props.onClick?.()}>
					{props.children}
				</button>
			</Match>
		</Switch>
	);
}
