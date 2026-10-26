import * as stylex from "@stylexjs/stylex";
import { attrs } from "../../utils";
import { pageStyles } from "../../styles/pageStyles";
import type { JSX } from "solid-js";

const styles = stylex.create({
	row: {
		display: "flex",
		alignItems: "center",
		justifyContent: "space-between",
		flexWrap: "wrap",
		gap: "16px",
		paddingInline: "16px",
	},
	text: { display: "flex", flexDirection: "column", gap: "6px" },
});

export function SettingsIntro(props: { children?: JSX.Element }) {
	return (
		<div {...attrs(stylex.props(styles.row))}>
			<div {...attrs(stylex.props(styles.text))}>
				<h1 {...attrs(stylex.props(pageStyles.heading))}>Ustawienia</h1>
				<p {...attrs(stylex.props(pageStyles.muted))}>
					Wybierz grupę, zarządzaj synchronizacją i dostosuj zajęcia w planie.
				</p>
			</div>
			{props.children}
		</div>
	);
}
