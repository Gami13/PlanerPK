import * as stylex from "@stylexjs/stylex";
import type { GetUserInfoResponse } from "../gen/pkplaner/v1/pkplaner_pb.js";
import { colors } from "../styles/tokens.stylex";
import { attrs } from "../utils";
import { PillTag } from "./PillTag";

type UserCardProps = {
	userInfo: GetUserInfoResponse;
};

const styles = stylex.create({
	card: {
		borderColor: colors.defaultBorder,
		borderRadius: "999px",
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.5rem",
		paddingBlock: "0.5rem",
		paddingInline: "0.75rem",
		alignItems: "center",
		backgroundColor: "#fff",
		display: "flex",
		maxWidth: "540px",
		width: "100%",
	},
	avatar: {
		borderRadius: "50%",
		flexShrink: 0,
		marginInlineEnd: "0.5rem",
		objectFit: "cover",
		height: "64px",
		width: "64px",
	},
	details: {
		display: "flex",
		flexDirection: "column",
	},
	name: {
		color: colors.primaryText,
		fontSize: "1rem",
		fontWeight: 600,
	},
	email: {
		color: colors.secondaryText,
		fontSize: "1rem",
	},
	metaRow: {
		gap: "0.5rem",
		display: "flex",
		flexWrap: "wrap",
	},
});

export function UserCard(props: UserCardProps) {
	return (
		<article {...attrs(stylex.props(styles.card))}>
			<img
				{...attrs(stylex.props(styles.avatar))}
				src={props.userInfo.picture}
				alt={`${props.userInfo.name}'s profile picture`}
			/>

			<div {...attrs(stylex.props(styles.details))}>
				<p {...attrs(stylex.props(styles.name))}>{props.userInfo.name}</p>

				<p {...attrs(stylex.props(styles.email))}>{props.userInfo.email}</p>

				<div {...attrs(stylex.props(styles.metaRow))}>
					<PillTag
						isActivable={true}
						isActive={props.userInfo.isActive}
						text={props.userInfo.isActive ? "Aktywny" : "Nieaktywny"}
					/>
					<PillTag
						isActivable={false}
						text={props.userInfo.chosenColumn === 0n ? "Nie wybrano kolumny" : `Kolumna ${props.userInfo.chosenColumn}`}
					/>
				</div>
			</div>
		</article>
	);
}
