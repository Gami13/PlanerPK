import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, createSignal, For, Show, useContext } from "solid-js";
import { AppContext } from "..";

import { rpc } from "../api_client";
import { Button } from "../Components/Button";
import { PageTitle } from "../Components/PageTitle";
import type { ClassInfo, ClassOverride, ColumnGroup } from "../gen/pkplaner/v1/pkplaner_pb.js";
import { attrs, getGroupByColumnIndex } from "../utils";

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
	// Override section styles
	overrideGrid: {
		gap: "0.75em",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	addOverrideRow: {
		gap: "0.75em",
		alignItems: "center",
		display: "flex",
		flexWrap: "wrap",
		width: "100%",
	},
	addOverrideSelect: {
		borderColor: "#cfd8e6",
		borderRadius: "10px",
		borderStyle: "solid",
		borderWidth: "1px",
		paddingBlock: "0.5em",
		paddingInline: "0.75em",
		color: "#1b2433",
		fontSize: "0.95em",
		maxWidth: "28em",
		width: "100%",
	},
	classCard: {
		borderColor: "#e0e6f0",
		borderRadius: "12px",
		borderStyle: "solid",
		borderWidth: "1px",
		gap: "0.75em",
		paddingBlock: "1em",
		paddingInline: "1.25em",
		transition: "border-color 0.15s ease, background-color 0.15s ease",
		backgroundColor: "#fafbfd",
		display: "flex",
		flexDirection: "column",
	},
	classCardActive: {
		borderColor: "#039be5",
		backgroundColor: "#f0f9ff",
	},
	classCardRow: {
		gap: "0.5em 1.5em",
		alignItems: "center",
		display: "flex",
		flexWrap: "wrap",
	},
	classTitle: {
		flex: "1",
		color: "#0f172a",
		fontSize: "0.95em",
		fontWeight: 600,
		minWidth: "10em",
	},
	classMeta: {
		color: "#546171",
		fontSize: "0.82em",
	},
	overrideControls: {
		gap: "0.75em 1.5em",
		alignItems: "flex-end",
		display: "flex",
		flexWrap: "wrap",
	},
	checkboxGroup: {
		gap: "0.25em",
		display: "flex",
		flexDirection: "column",
	},
	checkboxLabel: {
		gap: "0.4em",
		alignItems: "center",
		color: "#1b2433",
		cursor: "pointer",
		display: "flex",
		fontSize: "0.88em",
	},
	timeGroup: {
		gap: "0.25em",
		display: "flex",
		flexDirection: "column",
	},
	timeLabel: {
		color: "#546171",
		fontSize: "0.82em",
		fontWeight: 500,
	},
	timeInput: {
		borderColor: "#cfd8e6",
		borderRadius: "8px",
		borderStyle: "solid",
		borderWidth: "1px",
		paddingBlock: "0.4em",
		paddingInline: "0.6em",
		color: "#1b2433",
		fontSize: "0.9em",
		width: "7em",
	},
	saveRow: {
		gap: "0.75em",
		alignItems: "center",
		display: "flex",
		flexWrap: "wrap",
	},
	feedbackSuccess: {
		color: "#15803d",
		fontSize: "0.82em",
	},
	feedbackError: {
		color: "#b91c1c",
		fontSize: "0.82em",
	},
	emptyNote: {
		paddingBlock: "1em",
		paddingInline: "0",
		color: "#546171",
		fontSize: "0.9em",
		textAlign: "center",
	},
});

async function changeChosenColumn(newColumn: number): Promise<boolean> {
	try {
		await rpc.setChosenColumn({
			chosenColumn: newColumn,
		});
		return true;
	} catch {
		return false;
	}
}

async function changeUserActive(isActive: boolean): Promise<boolean> {
	try {
		await rpc.setUserActive({
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

async function fetchClasses(): Promise<ClassInfo[]> {
	try {
		const res = await rpc.getClasses({});
		// Sort by day then start_time for a stable order
		return [...res.classes].sort((a, b) => {
			const dayOrder = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];
			const da = dayOrder.indexOf(a.day);
			const db_ = dayOrder.indexOf(b.day);
			if (da !== db_) return da - db_;
			return a.startTime.localeCompare(b.startTime);
		});
	} catch {
		return [];
	}
}

async function fetchOverrides(): Promise<Map<string, ClassOverride>> {
	try {
		const res = await rpc.getOverrides({});
		const map = new Map<string, ClassOverride>();
		for (const o of res.overrides) {
			map.set(o.classHash, o);
		}
		return map;
	} catch {
		return new Map();
	}
}

const DAY_LABELS: Record<string, string> = {
	Monday: "Pon",
	Tuesday: "Wt",
	Wednesday: "Śr",
	Thursday: "Czw",
	Friday: "Pt",
	Saturday: "Sob",
	Sunday: "Ndz",
};

function ClassCard(props: {
	class: ClassInfo;
	override: ClassOverride | undefined;
	onSaved: (updated: ClassOverride) => void;
}) {
	const initial = () => props.override;

	const [isHidden, setIsHidden] = createSignal(initial()?.isHidden ?? false);
	const [isTransparent, setIsTransparent] = createSignal(initial()?.isTransparent ?? false);
	const [startHour, setStartHour] = createSignal(initial()?.startHour ?? "");
	const [status, setStatus] = createSignal<"idle" | "saving" | "ok" | "error">("idle");

	const hasActiveOverride = () => isHidden() || isTransparent() || startHour() !== "";

	async function save() {
		setStatus("saving");
		try {
			await rpc.setOverride({
				override: {
					classHash: props.class.hash,
					startHour: startHour(),
					isHidden: isHidden(),
					isTransparent: isTransparent(),
				},
			});
			setStatus("ok");
			props.onSaved({
				classHash: props.class.hash,
				startHour: startHour(),
				isHidden: isHidden(),
				isTransparent: isTransparent(),
			} as ClassOverride);
			setTimeout(() => setStatus("idle"), 2500);
		} catch {
			setStatus("error");
			setTimeout(() => setStatus("idle"), 3000);
		}
	}

	const cardStyle = () =>
		hasActiveOverride() ? stylex.props(styles.classCard, styles.classCardActive) : stylex.props(styles.classCard);

	return (
		<div {...attrs(cardStyle())}>
			{/* Header row */}
			<div {...attrs(stylex.props(styles.classCardRow))}>
				<span {...attrs(stylex.props(styles.classTitle))}>
					{props.class.name}
					<Show when={props.class.type}> ({props.class.type})</Show>
				</span>
				<span {...attrs(stylex.props(styles.classMeta))}>
					{DAY_LABELS[props.class.day] ?? props.class.day} &nbsp;·&nbsp;
					{props.class.startTime}–{props.class.endTime}
					<Show when={props.class.teacher}> &nbsp;·&nbsp; {props.class.teacher}</Show>
					<Show when={props.class.location}> &nbsp;·&nbsp; {props.class.location}</Show>
				</span>
			</div>

			{/* Controls */}
			<div {...attrs(stylex.props(styles.overrideControls))}>
				<div {...attrs(stylex.props(styles.checkboxGroup))}>
					<label {...attrs(stylex.props(styles.checkboxLabel))} id={`hidden-label-${props.class.hash}`}>
						<input
							id={`hidden-${props.class.hash}`}
							type="checkbox"
							checked={isHidden()}
							onChange={(e) => setIsHidden(e.currentTarget.checked)}
						/>
						Ukryj zajęcia
					</label>
					<label {...attrs(stylex.props(styles.checkboxLabel))} id={`transparent-label-${props.class.hash}`}>
						<input
							id={`transparent-${props.class.hash}`}
							type="checkbox"
							checked={isTransparent()}
							onChange={(e) => setIsTransparent(e.currentTarget.checked)}
						/>
						Pokaż poprzednią pozycję jako transparentną
					</label>
				</div>

				<Show when={!isHidden()}>
					<div {...attrs(stylex.props(styles.timeGroup))}>
						<label {...attrs(stylex.props(styles.timeLabel))} for={`start-hour-${props.class.hash}`}>
							Przesuń na godzinę
						</label>
						<input
							id={`start-hour-${props.class.hash}`}
							type="time"
							value={startHour()}
							onInput={(e) => setStartHour(e.currentTarget.value)}
							{...attrs(stylex.props(styles.timeInput))}
						/>
					</div>
				</Show>
			</div>

			{/* Save row */}
			<div {...attrs(stylex.props(styles.saveRow))}>
				<Button isPrimary={true} onClick={save}>
					{status() === "saving" ? "Zapisywanie…" : "Zapisz"}
				</Button>
				<Show when={status() === "ok"}>
					<span {...attrs(stylex.props(styles.feedbackSuccess))}>✓ Zapisano</span>
				</Show>
				<Show when={status() === "error"}>
					<span {...attrs(stylex.props(styles.feedbackError))}>✗ Błąd zapisu</span>
				</Show>
			</div>
		</div>
	);
}

export function Settings() {
	const AppState = useContext(AppContext);
	const [columns] = createResource(fetchColumns);
	const [classes, { refetch: refetchClasses }] = createResource(fetchClasses);
	const [overrides, { mutate: mutateOverrides }] = createResource(fetchOverrides);
	const [addedClassHashes, setAddedClassHashes] = createSignal<Set<string>>(new Set());
	const [selectedClassHash, setSelectedClassHash] = createSignal("");

	const chosenColumn = () => Number(AppState.store.userInfo?.chosenColumn ?? 0n);

	function handleOverrideSaved(updated: ClassOverride) {
		mutateOverrides((prev) => {
			const next = new Map(prev ?? []);
			if (!updated.isHidden && !updated.isTransparent && updated.startHour === "") {
				next.delete(updated.classHash);
				setAddedClassHashes((old) => {
					const newSet = new Set(old);
					newSet.delete(updated.classHash);
					return newSet;
				});
			} else {
				next.set(updated.classHash, updated);
			}
			return next;
		});
	}

	const visibleClasses = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = addedClassHashes();
		return classesValue.filter((cls) => overridesValue?.has(cls.hash) || added.has(cls.hash));
	});

	const classesAvailableToAdd = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = addedClassHashes();
		return classesValue.filter((cls) => !overridesValue?.has(cls.hash) && !added.has(cls.hash));
	});

	function addSelectedClassToOverrides() {
		const classHash = selectedClassHash();
		if (!classHash) return;
		setAddedClassHashes((old) => {
			const newSet = new Set(old);
			newSet.add(classHash);
			return newSet;
		});
		setSelectedClassHash("");
	}

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
						onClick={() => {
							if (!AppState.store.userInfo) {
								return;
							}

							const newStatus = !AppState.store.userInfo.isActive;
							changeUserActive(newStatus).then((success) => {
								if (!success) {
									alert("Wystąpił błąd podczas zmiany statusu konta. Spróbuj ponownie.");
									return;
								}

								AppState.setStore("userInfo", "isActive", newStatus);
								alert(`Konto oznaczono jako ${newStatus ? "aktywne" : "nieaktywne"}.`);
							});
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
								refetchClasses();
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
						{chosenColumn() === 0
							? "Nie wybrano kolumny"
							: `${chosenColumn() ? getGroupByColumnIndex(columns(), chosenColumn())?.label : "Ładowanie..."}`}
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
												<option value={option.column} selected={chosenColumn() === option.column}>
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

				{/* ── Overrides section ────────────────────────────────── */}
				<article {...attrs(stylex.props(styles.setting))}>
					<h2 {...attrs(stylex.props(styles.settingHeader))}>Nadpisania zajęć</h2>
					<p {...attrs(stylex.props(styles.settingDescription))}>
						Możesz ukryć zajęcia, przesunąć ich godzinę lub zostawić poprzednią pozycję jako transparentną w kalendarzu.
						Zmiany zostaną zastosowane przy następnej synchronizacji.
					</p>

					<Show when={chosenColumn() <= 0}>
						<p {...attrs(stylex.props(styles.emptyNote))}>Wybierz kolumnę powyżej, aby zobaczyć listę zajęć.</p>
					</Show>

					<Show when={chosenColumn() > 0}>
						<Show when={classes.loading}>
							<p {...attrs(stylex.props(styles.emptyNote))}>Ładowanie zajęć…</p>
						</Show>
						<Show when={classes.error}>
							<p {...attrs(stylex.props(styles.emptyNote))}>Błąd ładowania zajęć.</p>
						</Show>
						<Show when={classes() && classes()!.length === 0}>
							<p {...attrs(stylex.props(styles.emptyNote))}>Brak zajęć dla wybranej kolumny.</p>
						</Show>
						<Show when={classes() && classes()!.length > 0}>
							<div {...attrs(stylex.props(styles.addOverrideRow))}>
								<select
									id="override-class-select"
									value={selectedClassHash()}
									onChange={(e) => setSelectedClassHash(e.currentTarget.value)}
									{...attrs(stylex.props(styles.addOverrideSelect))}
								>
									<option value="">Wybierz zajęcia do nadpisania…</option>
									<For each={classesAvailableToAdd()}>
										{(cls) => (
											<option value={cls.hash}>
												{cls.name}
												<Show when={cls.type}> ({cls.type})</Show>
												{" - "}
												{DAY_LABELS[cls.day] ?? cls.day} {cls.startTime}-{cls.endTime} -- {cls.teacher} @ {cls.location}
											</option>
										)}
									</For>
								</select>
								<Button isPrimary={false} onClick={addSelectedClassToOverrides}>
									Dodaj nadpisanie
								</Button>
							</div>

							<div {...attrs(stylex.props(styles.overrideGrid))}>
								<For each={visibleClasses()}>
									{(cls) => (
										<ClassCard class={cls} override={overrides()?.get(cls.hash)} onSaved={handleOverrideSaved} />
									)}
								</For>
								<Show when={visibleClasses().length === 0}>
									<p {...attrs(stylex.props(styles.emptyNote))}>
										Brak aktywnych nadpisań. Wybierz zajęcia z listy powyżej.
									</p>
								</Show>
							</div>
						</Show>
					</Show>
				</article>
			</section>
		</main>
	);
}
