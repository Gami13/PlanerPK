import { useLocation, useNavigate } from "@solidjs/router";
import { onMount, useContext } from "solid-js";
import { PageTitle } from "../Components/PageTitle";
import * as stylex from "@stylexjs/stylex";
import { AppHeader } from "../Components/AppHeader";
import { pageStyles } from "../styles/pageStyles";
import { attrs } from "../utils";
import { AppContext, updateLoggedInUser } from "../appContext";
import { authorizeByCode } from "../services/pkplaner";
import { setAccessToken } from "../utils";

export function Authorize() {
	const location = useLocation();
	const navigate = useNavigate();
	const AppState = useContext(AppContext);

	onMount(async () => {
		const code = new URLSearchParams(location.search).get("code");
		if (!code) {
			return;
		}

		try {
			const data = await authorizeByCode(code);
			if (!data.accessToken) {
				return;
			}

			setAccessToken(data.accessToken);
			await updateLoggedInUser(AppState.setStore);
			navigate("/settings");
		} catch (error) {
			console.error("Authorization error:", error);
		}
	});

	return (
		<main {...attrs(stylex.props(pageStyles.page))}>
			<PageTitle title="Autoryzacja" />
			<AppHeader />
			<div {...attrs(stylex.props(pageStyles.intro))} role="status">
				<h1 {...attrs(stylex.props(pageStyles.heading))}>Łączenie konta…</h1>
				<p {...attrs(stylex.props(pageStyles.muted))}>Za chwilę przejdziesz do ustawień swojego planu.</p>
			</div>
		</main>
	);
}
