import * as stylex from "@stylexjs/stylex";
import { attrs } from "../../utils";
import { pageStyles } from "../../styles/pageStyles";
import type { JSX } from "solid-js";
import { A } from "@solidjs/router";

const styles = stylex.create({
	row: {
		display: "flex",
		alignItems: "flex-start",
		justifyContent: "space-between",
		flexWrap: "wrap",
		gap: "16px",
		paddingInline: "16px",
	},
	text: { display: "flex", flexDirection: "column", gap: "8px", position: "relative" },
	back: {
		position: "absolute",
		bottom: "calc(100% + 5px)",
		left: 0,
		fontSize: "0.75rem",
		lineHeight: 1,
		color: "#576475",
		textDecoration: "none",
	},
	description: { lineHeight: 1.5 },
});

export function SettingsIntro(props: { children?: JSX.Element }) {
	return (
		<div {...attrs(stylex.props(styles.row))}>
			<div {...attrs(stylex.props(styles.text))}>
				<A href="/" {...attrs(stylex.props(styles.back))}>← Plan zajęć</A>
				<h1 {...attrs(stylex.props(pageStyles.heading))}>Ustawienia</h1>
				<p {...attrs(stylex.props(pageStyles.muted, styles.description))}>
					Wybierz grupę, zarządzaj synchronizacją i dostosuj zajęcia w planie.
				</p>
			</div>
			{props.children}
		</div>
	);
}
