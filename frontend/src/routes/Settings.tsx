import * as stylex from "@stylexjs/stylex";
import { createResource, For, Show, useContext } from "solid-js";
import { AppContext } from "..";

import { rpc } from "../api_client";
import { Button } from "../Components/Button";
import { PageTitle } from "../Components/PageTitle";
import type { ColumnGroup } from "../gen/v1/pkplaner_pb.js";
import { attrs, getAccessToken, getGroupByColumnIndex } from "../utils";

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

async function changeChosenColumn(newColumn: number): Promise<boolean> {
	try {
		await rpc.setChosenColumn({
			chosenColumn: newColumn,
			accessToken: getAccessToken() || "",
		});
		return true;
	} catch {
		return false;
	}
}

async function changeUserActive(isActive: boolean): Promise<boolean> {
	const accessToken = getAccessToken();
	if (!accessToken) {
		return false;
	}

	try {
		await rpc.setUserActive({
			accessToken,
			isActive,
		});
		return true;
	} catch {
		return false;
	}
}

async function fetchColumns(): Promise<ColumnGroup[]> {
	const res = await rpc.getColumns({});
	return res.columns;
}

/** Group columns by category for optgroup rendering. */
function groupByCategory(columns: ColumnGroup[]): Map<string, ColumnGroup[]> {
	const map = new Map<string, ColumnGroup[]>();
	for (const col of columns) {
		const key = col.category || "";
		if (!map.has(key)) map.set(key, []);
		map.get(key)!.push(col);
	}
	return map;
}

export function Settings() {
	const AppState = useContext(AppContext);
	const [columns] = createResource(fetchColumns);

	return (
		<main {...attrs(stylex.props(styles.main))}>
			<PageTitle title="Ustawienia" />

			<section {...attrs(stylex.props(styles.basicSection))}>
				<div {...attrs(stylex.props(styles.description))}>
					<h1 {...attrs(stylex.props(styles.header))}>Ustawienia konta</h1>
					<p>Tu możesz zmienić ustawienia swojego konta.</p>
				</div>

				<article {...attrs(stylex.props(styles.setting))}>
					<h2 {...attrs(stylex.props(styles.settingHeader))}>Status aktywności</h2>
					<p {...attrs(stylex.props(styles.settingDescription))}>
						Obecny status: {AppState.store.userInfo?.isActive ? "Aktywny" : "Nieaktywny"}
					</p>
					<Button
						isPrimary={true}
						onClick={async () => {
							if (!AppState.store.userInfo) {
								return;
							}

							const newStatus = !AppState.store.userInfo.isActive;
							const success = await changeUserActive(newStatus);

							if (!success) {
								alert("Wystąpił błąd podczas zmiany statusu konta. Spróbuj ponownie.");
								return;
							}

							AppState.setStore("userInfo", "isActive", newStatus);
							alert(`Konto oznaczono jako ${newStatus ? "aktywne" : "nieaktywne"}.`);
						}}
					>
						{AppState.store.userInfo?.isActive ? "Oznacz konto jako nieaktywne" : "Oznacz konto jako aktywne"}
					</Button>
				</article>
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

								AppState.setStore("userInfo", "chosenColumn", BigInt(chosenColumn as string));
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
						{AppState.store.userInfo?.chosenColumn === 0n
							? "Nie wybrano kolumny"
							: `${AppState.store.userInfo?.chosenColumn ? getGroupByColumnIndex(columns(), Number(AppState.store.userInfo.chosenColumn))?.label : "Ładowanie..."}`}
					</p>
					<select {...attrs(stylex.props(styles.columnSelect))} name="chosen_column" id="chosen_column">
						<Show when={columns.error}>
							<option disabled>Błąd ładowania kolumn</option>
						</Show>
						<Show when={columns.loading}>
							<option disabled>Ładowanie...</option>
						</Show>
						<Show when={columns()}>
							<For each={[...groupByCategory(columns()!)]}>
								{([category, options]) => (
									<optgroup label={category || "Inne"}>
										<For each={options}>
											{(option) => (
												<option
													value={option.column}
													selected={Number(AppState.store.userInfo?.chosenColumn ?? 0n) === option.column}
												>
													{option.label}
												</option>
											)}
										</For>
									</optgroup>
								)}
							</For>
						</Show>
					</select>
					<Button isPrimary={true}>Zmień kolumnę</Button>
				</form>
			</section>
		</main>
	);
}
