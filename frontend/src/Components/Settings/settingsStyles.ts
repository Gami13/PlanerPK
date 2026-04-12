import * as stylex from "@stylexjs/stylex";
import { borderRadius, colors, spacing } from "../../styles/tokens.stylex";

export const inputBase = {
	borderColor: colors.defaultBorder,
	borderStyle: "solid",
	borderWidth: "1px",
	color: colors.primaryText,
	fontSize: "1em",
};

export const flexRow = {
	gap: "0.75em",
	alignItems: "center",
	display: "flex",
	flexWrap: "wrap",
};

export const flexColumn = {
	gap: "0.75em",
	display: "flex",
	flexDirection: "column",
};

export const settingsStyles = stylex.create({
	columnSelect: {
		...inputBase,
		borderRadius: borderRadius.lg,
		paddingBlock: "0.6em",
		paddingInline: "0.9em",
		maxWidth: "20em",
		width: "100%",
	},
	overrideGrid: {
		...flexColumn,
		width: "100%",
	},
	addOverrideRow: {
		...flexRow,
		width: "100%",
		justifyContent: "space-between",
	},
	addOverrideSelect: {
		flex: 1,
		maxWidth: "28em",
	},
	classCard: {
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.lg,
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		paddingBlock: spacing.md,
		paddingInline: "1.25em",
		transition: "border-color 0.15s ease, background-color 0.15s ease",
		backgroundColor: colors.lightCardBg,
		display: "flex",
		flexDirection: "column",
	},
	classCardActive: {
		borderColor: colors.primary,
		backgroundColor: colors.primaryLightBg,
	},
	classCardRow: {
		gap: "0.5em 1.5em",
		alignItems: "center",
		display: "flex",
		flexWrap: "wrap",
	},
	classTitle: {
		flex: "1",
		margin: 0,
		color: colors.primaryText,
		fontSize: "0.95em",
		fontWeight: 600,
		minWidth: "10em",
	},
	classMeta: {
		margin: 0,
		color: colors.secondaryText,
		fontSize: "0.82em",
	},
	overrideControls: {
		...flexRow,
		alignItems: "flex-end",
		columnGap: "1.5em",
		rowGap: "0.75em",
	},
	checkboxGroup: {
		...flexColumn,
	},
	checkboxLabel: {
		gap: "0.4em",
		alignItems: "center",
		color: colors.primaryText,
		cursor: "pointer",
		display: "flex",
		fontSize: "0.88em",
	},
	timeGroup: {
		...flexColumn,
		marginLeft: "auto",
	},
	timeLabel: {
		color: colors.secondaryText,
		fontSize: "0.82em",
		fontWeight: 500,
	},
	timeInput: {
		...inputBase,
		borderRadius: borderRadius.sm,
		paddingBlock: "0.4em",
		paddingInline: "0.6em",
		fontSize: "0.9em",
		width: "12em",
	},
	saveRow: {
		...flexRow,
	},
	feedbackSuccess: {
		color: colors.activeText,
		fontSize: "0.82em",
	},
	feedbackError: {
		color: colors.inactiveText,
		fontSize: "0.82em",
	},
	emptyNote: {
		paddingBlock: spacing.md,
		paddingInline: 0,
		color: colors.secondaryText,
		fontSize: "0.9em",
		textAlign: "center",
	},
});
