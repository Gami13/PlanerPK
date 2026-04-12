import { Button } from "../../Components/Button";
import { UIText } from "../../constants";
import { SettingCard } from "./SettingCard";

type AccountStatusCardProps = {
	isActive: boolean | undefined;
	onToggle: () => void;
};

export function AccountStatusCard(props: AccountStatusCardProps) {
	return (
		<SettingCard
			title={UIText.settings.account.title}
			description={props.isActive ? UIText.settings.account.active : UIText.settings.account.inactive}
		>
			<Button isPrimary={true} onClick={props.onToggle}>
				{props.isActive ? UIText.settings.account.markInactive : UIText.settings.account.markActive}
			</Button>
		</SettingCard>
	);
}
