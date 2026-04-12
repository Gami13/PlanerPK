import * as stylex from "@stylexjs/stylex";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";

export function SettingsIntro() {
	return (
		<div {...attrs(stylex.props(settingsStyles.description))}>
			<h1 {...attrs(stylex.props(settingsStyles.header))}>Ustawienia konta</h1>
			<p>Tu możesz zmienić ustawienia swojego konta.</p>
		</div>
	);
}
