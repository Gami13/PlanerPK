import * as stylex from "@stylexjs/stylex";

// Colors
export const colors = stylex.defineConsts({
	defaultBorder: "#e4e8ee",
	secondaryText: "#546171",
	primaryText: "#0f172a",
	activeBackground: "#dcfce7",
	activeText: "#15803d",
	inactiveBackground: "#fee2e2",
	inactiveText: "#b91c1c",
	primary: "#039be5",
	white: "#fff",
	lightBg: "#f4f6fb",
	lightCardBg: "#fafbfd",
	primaryLightBg: "#f0f9ff",
});

// Spacing - consistent scale
export const spacing = stylex.defineConsts({
	xs: "0.25rem",
	sm: "0.5rem",
	md: "1rem",
	lg: "1.5rem",
	xl: "2rem",
});

// Typography
export const typography = stylex.create({
	xs: { fontSize: "0.75rem" },
	sm: { fontSize: "0.82rem" },
	base: { fontSize: "1em" },
	md: { fontSize: "0.95em" },
	lg: { fontSize: "1.4em" },
	xl: { fontSize: "2.8em" },
	xsWeight: { fontWeight: 400 },
	smWeight: { fontWeight: 500 },
	mdWeight: { fontWeight: 600 },
});

// Border radius
export const borderRadius = stylex.defineConsts({
	none: "0",
	sm: "8px",
	md: "10px",
	lg: "12px",
	xl: "16px",
	full: "9999px",
});

// Breakpoints - reference values (must use literal strings in StyleX objects)
export const breakpoints = stylex.defineConsts({
	mobile: "@media (max-width: 33.75em)",
	tablet: "@media (min-width: 38em)",
	desktop: "@media (min-width: 48em)",
});
