import * as stylex from "@stylexjs/stylex";
import { breakpoints, spacing } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";

const style = stylex.create({
	description: {
		gap: spacing.sm,
		display: "flex",
		flexDirection: "column",
		textAlign: {
			[breakpoints.desktop]: "left",
			default: "center",
		},
		width: "min(100%, 48em)",
	},
	header: {
		fontSize: "2.8em",
		fontWeight: 500,
	},
});
export function SettingsIntro() {
	return (
		<div {...attrs(stylex.props(style.description))}>
			<h1 {...attrs(stylex.props(style.header))}>Ustawienia konta</h1>
			<p>Tu możesz zmienić ustawienia swojego konta.</p>
		</div>
	);
}
