import * as stylex from "@stylexjs/stylex";
import { Button } from "../../Components/Button";
import { UI_MESSAGES } from "../../constants";
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
				Obecny status: {props.isActive ? UI_MESSAGES.ACCOUNT_ACTIVE : UI_MESSAGES.ACCOUNT_INACTIVE}
			</p>
			<Button isPrimary={true} onClick={props.onToggle}>
				{props.isActive ? UI_MESSAGES.MARK_INACTIVE : UI_MESSAGES.MARK_ACTIVE}
			</Button>
		</article>
	);
}
