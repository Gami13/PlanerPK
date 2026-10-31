import * as stylex from "@stylexjs/stylex";
import { colors } from "./tokens.stylex";

export const pageStyles = stylex.create({
	page: {
		width: "100%",
		maxWidth: "84rem",
		marginInline: "auto",
		paddingTop: "clamp(40px, 5vw, 48px)",
		paddingInline: "clamp(12px, 3vw, 32px)",
		paddingBottom: "clamp(24px, 4vw, 40px)",
		display: "flex",
		flexDirection: "column",
		gap: "24px",
		color: colors.primaryText,
	},
	section: { display: "flex", flexDirection: "column", gap: "20px", width: "100%", minWidth: 0 },
	heading: { margin: 0, fontSize: "1.65rem", fontWeight: 600, letterSpacing: "-0.025em", lineHeight: 1.25 },
	intro: { display: "flex", flexDirection: "column", gap: "6px", paddingInline: "16px" },
	muted: { color: colors.secondaryText, fontSize: "0.9rem", lineHeight: 1.6 },
	footer: { paddingInline: "16px", color: colors.secondaryText, fontSize: "0.8rem" },
});
