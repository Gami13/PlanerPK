import * as stylex from "@stylexjs/stylex";
import type { GetUserInfoResponse } from "../gen/pkplaner/v1/pkplaner_pb.js";
import { colors } from "../styles/tokens.stylex";
import { attrs } from "../utils";
import { PillTag } from "./PillTag";

type UserCardProps = {
	userInfo: GetUserInfoResponse;
	groupLabel?: string;
	compact?: boolean;
};

const styles = stylex.create({
	card: {
		borderColor: colors.defaultBorder,
		borderRadius: "6px",
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "12px",
		paddingBlock: "12px",
		paddingInline: "14px",
		alignItems: "center",
		backgroundColor: "#fff",
		display: "flex",
		minWidth: 0,
		flex: "1 1 auto",
		width: "100%",
	},
	avatar: {
		borderRadius: "50%",
		flexShrink: 0,
		objectFit: "cover",
		height: "40px",
		width: "40px",
	},
	compact: { borderWidth: 0, paddingInline: 0, paddingBlock: 0, paddingTop: "4px", width: "auto", flex: "0 1 auto", maxWidth: "100%" },
	details: {
		minWidth: 0,
		gap: "2px",
		display: "flex",
		flexDirection: "column",
	},
	name: {
		color: colors.primaryText,
		fontSize: "0.9rem",
		fontWeight: 600,
	},
	identity: { display: "flex", alignItems: "baseline", flexWrap: "wrap", gap: "2px 8px", minWidth: 0 },
	email: {
		color: colors.secondaryText,
		fontSize: "0.8rem",
		overflow: "hidden",
		textOverflow: "ellipsis",
		whiteSpace: "nowrap",
	},
	metaRow: {
		gap: "0.5rem",
		display: "flex",
		flexWrap: "wrap",
	},
});

export function UserCard(props: UserCardProps) {
	return (
		<article {...attrs(stylex.props(styles.card, props.compact && styles.compact))}>
			<img {...attrs(stylex.props(styles.avatar))} src={props.userInfo.picture} alt="" referrerpolicy="no-referrer" />

			<div {...attrs(stylex.props(styles.details))}>
				<div {...attrs(stylex.props(styles.identity))}>
				<p {...attrs(stylex.props(styles.name))}>{props.userInfo.name}</p>

				<p {...attrs(stylex.props(styles.email))}>{props.userInfo.email}</p>
				</div>

				<div {...attrs(stylex.props(styles.metaRow))}>
					<PillTag
						isActivable={true}
						isActive={props.userInfo.isActive}
						text={props.userInfo.isActive ? "Synchronizacja włączona" : "Synchronizacja wyłączona"}
					/>
					<PillTag
						isActivable={false}
						text={props.userInfo.chosenColumn === 0n ? "Nie wybrano grupy" : (props.groupLabel ?? "Ładowanie grupy…")}
					/>
				</div>
			</div>
		</article>
	);
}
