import * as stylex from "@stylexjs/stylex";
import { borderRadius, breakpoints, colors, spacing } from "../../styles/tokens.stylex";

const inputBase = {
	borderColor: colors.defaultBorder,
	borderStyle: "solid",
	borderWidth: "1px",
	color: colors.primaryText,
	fontSize: "1em",
};

const flexRow = {
	gap: "0.75em",
	alignItems: "center",
	display: "flex",
	flexWrap: "wrap",
};

const flexColumn = {
	gap: "0.75em",
	display: "flex",
	flexDirection: "column",
};

export const settingsStyles = stylex.create({
	main: {
		padding: spacing.xl,
		gap: spacing.xl,
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	basicSection: {
		gap: spacing.md,
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
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
	setting: {
		padding: spacing.lg,
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.xl,
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		backgroundColor: colors.white,
		display: "flex",
		flexDirection: "column",
		width: "min(100%, 48em)",
	},
	settingHeader: {
		margin: 0,
		fontSize: "1.4em",
		fontWeight: 600,
	},
	settingDescription: {
		margin: 0,
		color: colors.secondaryText,
	},
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
	},
	addOverrideSelect: {
		...inputBase,
		borderRadius: borderRadius.md,
		paddingBlock: spacing.sm,
		paddingInline: "0.75em",
		fontSize: "0.95em",
		maxWidth: "28em",
		width: "100%",
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
		color: colors.primaryText,
		fontSize: "0.95em",
		fontWeight: 600,
		minWidth: "10em",
	},
	classMeta: {
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
		width: "7em",
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
