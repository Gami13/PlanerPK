import * as stylex from "@stylexjs/stylex";
import { borderRadius, colors, spacing } from "./tokens.stylex";

export const primitiveStyles = stylex.create({
	buttonBase: {
		borderRadius: borderRadius.full,
		borderStyle: "none",
		borderWidth: 0,
		paddingBlock: "0.45em",
		paddingInline: spacing.lg,
		textDecoration: "none",
		transition: "transform 0.15s ease, box-shadow 0.15s ease",
		alignItems: "center",
		cursor: "pointer",
		display: "inline-flex",
		fontSize: "1em",
		fontWeight: 500,
		justifyContent: "center",
		textAlign: "center",
	},
	buttonPrimary: {
		backgroundColor: colors.primary,
		color: colors.white,
	},
	buttonSecondary: {
		borderColor: colors.defaultBorder,
		borderStyle: "solid",
		borderWidth: "1px",
		backgroundColor: colors.white,
		color: colors.primaryText,
	},
	pillBase: {
		borderRadius: borderRadius.full,
		paddingBlock: spacing.xs,
		paddingInline: spacing.md,
		backgroundColor: colors.lightBg,
		color: colors.secondaryText,
		fontSize: "0.8rem",
	},
	pillStatusText: {
		fontWeight: 600,
	},
	pillStatusActive: {
		backgroundColor: colors.activeBackground,
		color: colors.activeText,
	},
	pillStatusInactive: {
		backgroundColor: colors.inactiveBackground,
		color: colors.inactiveText,
	},
});
