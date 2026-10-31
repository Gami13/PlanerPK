import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { Show, splitProps, type JSX } from "solid-js";
import { primitiveStyles } from "../styles/primitives";
import { attrs } from "../utils";

type ButtonProps = JSX.ButtonHTMLAttributes<HTMLButtonElement> & {
	isPrimary?: boolean;
	styles?: stylex.StyleXStyles;
	href?: string;
	isVanillaLink?: boolean;
};

export function Button(props: ButtonProps) {
	const [, buttonProps] = splitProps(props, ["children", "isPrimary", "styles", "href", "isVanillaLink", "class"]);
	const buttonAttrs = () => {
		const styled = attrs(
			stylex.props(
				primitiveStyles.buttonBase,
				props.isPrimary ? primitiveStyles.buttonPrimary : primitiveStyles.buttonSecondary,
				props.styles
			)
		);
		return { ...styled, class: [styled.class, props.class].filter(Boolean).join(" ") };
	};

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
				<button {...buttonAttrs()} {...buttonProps}>
					{props.children}
				</button>
			</Show>
		</>
	);
}
