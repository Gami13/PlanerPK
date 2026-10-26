import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { Show, useContext } from "solid-js";
import { AppContext } from "../appContext";
import { attrs } from "../utils";

const styles = stylex.create({
	header: {
		display: "flex",
		alignItems: "center",
		justifyContent: "space-between",
		gap: "16px",
		padding: "0 16px 18px",
		borderBottomWidth: "1px",
		borderBottomStyle: "solid",
		borderBottomColor: "#dfe5ed",
	},
	brand: {
		display: "inline-flex",
		alignItems: "center",
		gap: "9px",
		color: "#0b2a45",
		fontWeight: 600,
		fontSize: "1rem",
		textDecoration: "none",
	},
	mark: { color: "#4b99d2", flexShrink: 0 },
	nav: { display: "flex", gap: "4px", flexWrap: "wrap" },
	link: {
		padding: "6px 10px",
		borderRadius: "4px",
		color: "#576475",
		fontSize: "0.85rem",
		textDecoration: "none",
		backgroundColor: { default: "transparent", ":hover": "#eef4fc" },
	},
	active: { backgroundColor: "#e4eefb", color: "#245a85", fontWeight: 500 },
});

export function AppHeader() {
	const app = useContext(AppContext);
	return (
		<header {...attrs(stylex.props(styles.header))}>
			<A href="/" {...attrs(stylex.props(styles.brand))}>
				<svg
					{...attrs(stylex.props(styles.mark))}
					width="22"
					height="22"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="1.8"
					aria-hidden="true"
				>
					<rect x="3" y="5" width="18" height="16" rx="3" />
					<path d="M7 3v4m10-4v4M3 11h18M8 15h3m2 0h3m-8 3h3" />
				</svg>
				PlanerPK
			</A>
			<nav {...attrs(stylex.props(styles.nav))} aria-label="Nawigacja główna">
				<A href="/" end {...attrs(stylex.props(styles.link))} activeClass={stylex.props(styles.active).className}>
					Plan zajęć
				</A>
				<Show when={app.store.userInfo}>
					<A href="/settings" {...attrs(stylex.props(styles.link))} activeClass={stylex.props(styles.active).className}>
						Ustawienia
					</A>
				</Show>
			</nav>
		</header>
	);
}
