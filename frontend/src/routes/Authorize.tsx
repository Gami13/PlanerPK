import { useLocation, useNavigate } from "@solidjs/router";
import { onMount, useContext } from "solid-js";
import { PageTitle } from "../Components/PageTitle";
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
		<main>
			<PageTitle title="Autoryzacja" />
			<h1>Autoryzacja...</h1>
		</main>
	);
}
