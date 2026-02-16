import { useLocation, useNavigate } from "@solidjs/router";
import { createEffect, useContext } from "solid-js";
import { AppContext, updateLoggedInUser } from "..";
import { rpc } from "../api_client";
import { PageTitle } from "../Components/PageTitle";
import { setAccessToken } from "../utils";

export function Authorize() {
	const location = useLocation();
	const navigate = useNavigate();
	const AppState = useContext(AppContext);

	createEffect(() => {
		const code = new URLSearchParams(location.search).get("code");

		if (code) {
			rpc
				.authorize({ code })
				.then((data) => {
					if (data.accessToken) {
						setAccessToken(data.accessToken);

						updateLoggedInUser(AppState.store, AppState.setStore).then(() => {
							navigate("/settings");
						});
					}
				})
				.catch((error) => {
					console.error("Authorization error:", error);
				});
		}
	});
	return (
		<main>
			<PageTitle title="Autoryzacja" />
			<h1>Autoryzacja...</h1>
		</main>
	);
}
