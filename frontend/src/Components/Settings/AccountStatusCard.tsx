import * as stylex from "@stylexjs/stylex";
import { Button } from "../../Components/Button";
import { attrs } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type AccountStatusCardProps = {
	isActive: boolean | undefined;
	onToggle: () => void;
};

export function AccountStatusCard(props: AccountStatusCardProps) {
	return (
		<article {...attrs(stylex.props(settingsStyles.setting))}>
			<h2 {...attrs(stylex.props(settingsStyles.settingHeader))}>Status aktywności</h2>
			<p {...attrs(stylex.props(settingsStyles.settingDescription))}>
				Obecny status: {props.isActive ? "Aktywny" : "Nieaktywny"}
			</p>
			<Button isPrimary={true} onClick={props.onToggle}>
				{props.isActive ? "Oznacz konto jako nieaktywne" : "Oznacz konto jako aktywne"}
			</Button>
		</article>
	);
}
