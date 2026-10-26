import * as stylex from "@stylexjs/stylex";
import { borderRadius, colors, spacing } from "./tokens.stylex";

export const primitiveStyles = stylex.create({
	buttonBase: {
		borderRadius: "5px",
		borderStyle: "solid",
		borderWidth: "1px",
		paddingBlock: "8px",
		paddingInline: "14px",
		minHeight: "38px",
		textDecoration: "none",
		transition: "background-color 0.15s ease, border-color 0.15s ease",
		outline: { default: "none", ":focus-visible": "2px solid #4b99d2" },
		outlineOffset: "3px",
		alignItems: "center",
		cursor: "pointer",
		display: "inline-flex",
		fontSize: "0.85rem",
		fontWeight: 500,
		justifyContent: "center",
		textAlign: "center",
	},
	buttonPrimary: {
		backgroundColor: { default: "#e4eefb", ":hover": "#d6e7fb" },
		borderColor: "#c9dcf5",
		color: "#245a85",
	},
	buttonSecondary: {
		borderColor: colors.defaultBorder,
		borderStyle: "solid",
		borderWidth: "1px",
		backgroundColor: { default: colors.white, ":hover": "#f4f7fb" },
		color: colors.secondaryText,
	},
	pillBase: {
		borderRadius: "3px",
		paddingBlock: spacing.xs,
		paddingInline: "6px",
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
