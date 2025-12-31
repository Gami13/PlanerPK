import * as stylex from "@stylexjs/stylex";
import { For, useContext } from "solid-js";
import { AppContext } from "..";

import { Button } from "../Components/Button";
import { PageTitle } from "../Components/PageTitle";
import { Endpoints } from "../constants";
import { attrs, getAccessToken } from "../utils";

const MEDIUM_SCREEN = "@media (min-width: 48em)" as const;
const styles = stylex.create({
	main: {
		padding: "2em",
		gap: "2em",
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	basicSection: {
		gap: "1em",
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	description: {
		gap: "0.5em",
		display: "flex",
		flexDirection: "column",
		textAlign: {
			[MEDIUM_SCREEN]: "left",
			default: "center",
		},
		width: "min(100%, 48em)",
	},
	header: {
		fontSize: "2.8em",
		fontWeight: 500,
	},
	setting: {
		padding: "1.5em",
		borderColor: "#e0e6f0",
		borderRadius: "16px",
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		backgroundColor: "#fff",
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
		color: "#4c576a",
	},
	columnSelect: {
		borderColor: "#cfd8e6",
		borderRadius: "12px",
		borderStyle: "solid",
		borderWidth: "1px",
		paddingBlock: "0.6em",
		paddingInline: "0.9em",
		color: "#1b2433",
		fontSize: "1em",
		maxWidth: "20em",
		width: "100%",
	},
});

async function changeChosenColumn(newColumn: number) {
	const response = await fetch(Endpoints.SetChosenColumn, {
		method: "PUT",
		headers: {
			"Content-Type": "application/json",
			Authorization: getAccessToken() || "",
		},
		body: JSON.stringify({ chosenColumn: newColumn }),
	});
	return response.ok;
}

const columnOptions = [
	{ value: 2, label: "Kolumna 2 (GK1)" },
	{ value: 3, label: "Kolumna 3 (GK2)" },
	{ value: 4, label: "Kolumna 4 (GK3)" },
	{ value: 5, label: "Kolumna 5 (GK4)" },
	{ value: 6, label: "Kolumna 6 (GK5)" },
	{ value: 7, label: "Kolumna 7 (GK6)" },
	{ value: 8, label: "Kolumna 8 (GK7)" },
	{ value: 9, label: "Kolumna 9 (GK8)" },
];

export function Settings() {
	const AppState = useContext(AppContext);

	return (
		<main {...attrs(stylex.props(styles.main))}>
			<PageTitle title="Ustawienia" />

			<section {...attrs(stylex.props(styles.basicSection))}>
				<div {...attrs(stylex.props(styles.description))}>
					<h1 {...attrs(stylex.props(styles.header))}>Ustawienia konta</h1>
					<p>Tu możesz zmienić ustawienia swojego konta.</p>
				</div>

				{/* <article {...attrs(stylex.props(styles.setting))}>
          			<h2 {...attrs(stylex.props(styles.settingHeader))}>Status aktywności</h2>
         			 	<p {...attrs(stylex.props(styles.settingDescription))}>Obecny status: {AppState.store.userInfo?.is_active ? 'Aktywny' : 'Nieaktywny'}</p>
          				<Button isPrimary={true}>Zmień status</Button>
        			</article> 
					*/}
				<form
					{...attrs(stylex.props(styles.setting))}
					method="post"
					onSubmit={async (e) => {
						e.preventDefault();
						const formData = new FormData(e.currentTarget);
						const chosenColumn = formData.get("chosen_column");
						if (chosenColumn) {
							const success = await changeChosenColumn(Number(chosenColumn));
							if (success) {
								alert("Pomyślnie zmieniono kolumnę.");

								AppState.setStore("userInfo", "chosen_column", Number(chosenColumn));
							} else {
								alert("Wystąpił błąd podczas zmiany kolumny. Spróbuj ponownie.");
							}
						}
					}}
				>
					<h2 {...attrs(stylex.props(styles.settingHeader))}>Wybrana kolumna</h2>
					<p>Zmianę kolumny wykonuj wyłącznie w uzasadnionych przypadkach, aby uniknąć ograniczeń konta.</p>
					<p {...attrs(stylex.props(styles.settingDescription))}>
						Obecna kolumna:{" "}
						{AppState.store.userInfo?.chosen_column === null
							? "Nie wybrano kolumny"
							: `Kolumna ${AppState.store.userInfo?.chosen_column}`}
					</p>
					<select {...attrs(stylex.props(styles.columnSelect))} name="chosen_column" id="chosen_column">
						<For each={columnOptions}>
							{(option) => (
								<option value={option.value} selected={AppState.store.userInfo?.chosen_column === option.value}>
									{option.label}
								</option>
							)}
						</For>
					</select>
					<Button isPrimary={true}>Zmień kolumnę</Button>
				</form>
			</section>
		</main>
	);
}
