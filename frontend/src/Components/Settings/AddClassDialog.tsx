import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, createSignal, createUniqueId, For, Show } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { getImportGroups, getPublicTimetable, setClassOverride } from "../../services/pkplaner";
import { Button } from "../Button";
import { ColumnOptions } from "../ColumnOptions";
import { createModalDialog } from "../createModalDialog";
import { attrs } from "../../utils";
import { DayLabels } from "../../constants";
import { overrideDialogStyles as shared } from "./ClassOverrideDialog";

const styles = stylex.create({
	dialog: { width: "min(calc(100% - 2rem), 38rem)" },
	filters: { display: "flex", flexDirection: "column", gap: "8px", paddingBlock: "8px" },
	input: {
		width: "100%",
		padding: "8px 10px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: "#dfe5ed",
		borderRadius: "4px",
		backgroundColor: "#fff",
		fontSize: "0.85rem",
		color: "#0b2a45",
	},
	list: { listStyle: "none", padding: 0, margin: 0, maxHeight: "min(45vh, 24rem)", overflowY: "auto" },
	row: {
		display: "flex",
		alignItems: "center",
		justifyContent: "space-between",
		gap: "12px",
		paddingBlock: "12px",
		borderBottomWidth: "1px",
		borderBottomStyle: "solid",
		borderBottomColor: "#eef2f7",
	},
	name: { fontSize: "0.9rem", fontWeight: 500, color: "#0b2a45" },
	status: { paddingBlock: "12px", fontSize: "0.85rem", color: "#576475" },
	disabled: { opacity: 0.55, cursor: "default", flexShrink: 0 },
});

export function AddClassDialog(props: {
	existingClasses: ClassInfo[];
	onAdded: (override: ClassOverride) => void;
	onClose: () => void;
}) {
	const id = createUniqueId();
	const [groups] = createResource(getImportGroups);
	const [selected, setSelected] = createSignal<number>();
	// Avoid suspending the page that owns the native dialog: detaching it drops
	// it out of the browser's modal top layer. Loading is rendered locally below.
	const availableGroups = () => (groups.error || groups.loading ? [] : (groups.latest ?? []));
	const source = () => selected() ?? availableGroups()[0]?.column;
	const [classes] = createResource(source, getPublicTimetable);
	const [search, setSearch] = createSignal("");
	const [pending, setPending] = createSignal<string>();
	const [error, setError] = createSignal("");
	const [added, setAdded] = createSignal(new Set<string>());
	const existing = createMemo(() => new Set([...props.existingClasses.map((classInfo) => classInfo.hash), ...added()]));
	const filtered = createMemo(() => {
		const query = search().trim().toLocaleLowerCase("pl");
		return (classes.error || classes.loading ? [] : (classes.latest ?? [])).filter((lesson) =>
			[lesson.name, lesson.teacher, lesson.type, lesson.location, lesson.recurrence, DayLabels[lesson.day]]
				.join(" ")
				.toLocaleLowerCase("pl")
				.includes(query)
		);
	});
	const dialog = createModalDialog(pending, () => props.onClose());
	async function add(lesson: ClassInfo) {
		const column = source();
		if (column === undefined || pending() || existing().has(lesson.hash)) return;
		setPending(lesson.hash);
		setError("");
		const override = {
			classHash: lesson.hash,
			sourceColumn: column,
			startHour: "",
			isHidden: false,
			isTransparent: false,
		} as ClassOverride;
		try {
			if (!(await setClassOverride(override))) {
				setError("Nie udało się dodać zajęć. Spróbuj ponownie.");
				return;
			}
			setAdded((previous) => new Set([...previous, lesson.hash]));
			props.onAdded(override);
		} finally {
			setPending(undefined);
		}
	}
	return (
		<dialog {...dialog.props} {...attrs(stylex.props(shared.dialog, styles.dialog))} aria-labelledby={`${id}-title`}>
			<header {...attrs(stylex.props(shared.header))}>
				<h2 id={`${id}-title`} {...attrs(stylex.props(shared.title))}>
					Dodaj zajęcia z innej grupy
				</h2>
				<p {...attrs(stylex.props(shared.meta))}>Wybierz grupę z tego samego roku i znajdź zajęcia.</p>
			</header>
			<div {...attrs(stylex.props(shared.body))}>
				<div {...attrs(stylex.props(styles.filters))}>
					<label for={`${id}-group`} {...attrs(stylex.props(shared.meta))}>
						Grupa
					</label>
					<select
						id={`${id}-group`}
						{...attrs(stylex.props(styles.input))}
						disabled={!!pending() || groups.loading}
						value={source() ?? ""}
						onChange={(event) => {
							setSelected(Number(event.currentTarget.value));
							setError("");
						}}
					>
						<ColumnOptions columns={availableGroups()} selected={source()} />
					</select>
					<label for={`${id}-search`} {...attrs(stylex.props(shared.meta))}>
						Szukaj zajęć
					</label>
					<input
						id={`${id}-search`}
						{...attrs(stylex.props(styles.input))}
						type="search"
						value={search()}
						onInput={(event) => setSearch(event.currentTarget.value)}
						placeholder="Nazwa, prowadzący, sala…"
					/>
				</div>
				<Show when={groups.loading || classes.loading}>
					<p {...attrs(stylex.props(styles.status))} role="status">
						Ładowanie zajęć…
					</p>
				</Show>
				<Show when={groups.error || classes.error}>
					<p {...attrs(stylex.props(shared.error))} role="alert">
						Nie udało się wczytać zajęć.
					</p>
				</Show>
				<Show when={!groups.loading && !groups.error && !availableGroups().length}>
					<p {...attrs(stylex.props(styles.status))}>Brak innych grup na tym roku.</p>
				</Show>
				<Show when={!classes.loading && !classes.error && !!source() && !filtered().length}>
					<p {...attrs(stylex.props(styles.status))}>Nie znaleziono zajęć.</p>
				</Show>
				<Show when={!classes.loading && !classes.error}>
					<ul {...attrs(stylex.props(styles.list))} aria-label="Zajęcia do dodania">
						<For each={filtered()}>
							{(lesson) => (
								<li {...attrs(stylex.props(styles.row))}>
									<div>
										<p {...attrs(stylex.props(styles.name))}>
											{lesson.name} {lesson.type ? `(${lesson.type})` : ""}
										</p>
										<p {...attrs(stylex.props(shared.meta))}>{lesson.teacher}</p>
										<p {...attrs(stylex.props(shared.meta))}>
											{DayLabels[lesson.day] ?? lesson.day} · {lesson.startTime}–{lesson.endTime}
											{lesson.location ? ` · ${lesson.location}` : ""}
											{lesson.recurrence ? ` · ${lesson.recurrence}` : ""}
										</p>
									</div>
									<Button
										isPrimary
										styles={(!!pending() || existing().has(lesson.hash)) && styles.disabled}
										aria-label={
											existing().has(lesson.hash)
												? `${lesson.name} — już w planie`
												: `Dodaj: ${lesson.name}, ${DayLabels[lesson.day] ?? lesson.day}, ${lesson.startTime}`
										}
										type="button"
										disabled={!!pending() || existing().has(lesson.hash)}
										onClick={() => void add(lesson)}
									>
										{existing().has(lesson.hash) ? "W planie" : pending() === lesson.hash ? "Dodawanie…" : "Dodaj"}
									</Button>
								</li>
							)}
						</For>
					</ul>
				</Show>
				<Show when={error()}>
					<p {...attrs(stylex.props(shared.error))} role="alert">
						{error()}
					</p>
				</Show>
				<p {...attrs(stylex.props(shared.meta, shared.hint))}>
					Dodane zajęcia możesz edytować lub usunąć, klikając je w planie.
				</p>
			</div>
			<footer {...attrs(stylex.props(shared.actions))}>
				<Button type="button" disabled={!!pending()} onClick={() => dialog.close()}>
					Gotowe
				</Button>
			</footer>
		</dialog>
	);
}
