import * as stylex from "@stylexjs/stylex";
import { colors } from "./tokens.stylex";

export const primitiveStyles = stylex.create({
	buttonBase: {
		borderRadius: "9999px",
		borderStyle: "none",
		borderWidth: 0,
		paddingBlock: "0.45em",
		paddingInline: "1.3em",
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
		color: "#fff",
	},
	buttonSecondary: {
		borderColor: colors.defaultBorder,
		borderStyle: "solid",
		borderWidth: "1px",
		backgroundColor: "#fff",
		color: colors.primaryText,
	},
	pillBase: {
		borderRadius: "999px",
		paddingBlock: "0.35rem",
		paddingInline: "0.75rem",
		backgroundColor: "#f4f6fb",
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
