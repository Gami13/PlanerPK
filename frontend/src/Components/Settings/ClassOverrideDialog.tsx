import * as stylex from "@stylexjs/stylex";
import { createSignal, createUniqueId, onMount, Show } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { setClassOverride } from "../../services/pkplaner";
import { colors } from "../../styles/tokens.stylex";
import { primitiveStyles } from "../../styles/primitives";
import { attrs } from "../../utils";
import { DayLabels } from "../../constants";
import { timetableStyles } from "../Timetable/timetableStyles";

const styles = stylex.create({
	dialog: {
		width: "min( calc(100% - 2rem), 30rem)",
		maxHeight: "calc(100dvh - 2rem)",
		overflowY: "auto",
		padding: 0,
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: "6px",
		backgroundColor: colors.white,
		color: colors.primaryText,
		boxShadow: "0 8px 32px rgba(15, 23, 42, 0.12)",
		"::backdrop": { backgroundColor: "rgba(30, 45, 65, 0.24)" },
	},
	form: { display: "flex", flexDirection: "column" },
	header: {
		position: "relative",
		display: "flex",
		flexDirection: "column",
		gap: "2px",
		padding: "12px 20px",
		backgroundColor: "#e4eefb",
		color: "#0b2a45",
	},
	headerTitle: { paddingRight: "22px" },
	headerFooter: { display: "flex", justifyContent: "space-between", flexWrap: "wrap", gap: "2px 12px", marginTop: "4px" },
	body: { display: "flex", flexDirection: "column", padding: "8px 20px 16px", gap: "4px" },
	title: { margin: 0, fontSize: "1.05rem", lineHeight: 1.3, fontWeight: 600 },
	meta: { margin: 0, fontSize: "0.8rem", color: "#576475", lineHeight: 1.45 },
	label: {
		display: "flex",
		alignItems: "center",
		justifyContent: "space-between",
		gap: "16px",
		fontSize: "0.85rem",
		paddingBlock: "12px",
		borderBottomWidth: "1px",
		borderBottomStyle: "solid",
		borderBottomColor: "#eef2f7",
	},
	copy: { display: "flex", flexDirection: "column", gap: "3px" },
	checkbox: { accentColor: "#4b99d2", flexShrink: 0 },
	hint: { marginTop: "10px" },
	time: {
		padding: "8px 10px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: "4px",
		fontSize: "0.85rem",
		backgroundColor: colors.white,
		color: colors.primaryText,
	},
	actions: {
		display: "flex",
		flexWrap: "wrap",
		justifyContent: "flex-end",
		gap: "8px",
		padding: "12px 20px",
		borderTopWidth: "1px",
		borderTopStyle: "solid",
		borderTopColor: "#dfe5ed",
	},
	reset: { marginRight: "auto" },
	error: { margin: 0, color: colors.inactiveText, fontSize: "0.85em" },
});

export function ClassOverrideDialog(props: {
	classInfo: ClassInfo;
	override?: ClassOverride;
	onSaved: (override: ClassOverride) => void;
	onClose: () => void;
}) {
	let dialog!: HTMLDialogElement;
	const id = createUniqueId();
	const [hidden, setHidden] = createSignal(props.override?.isHidden ?? false);
	const [transparent, setTransparent] = createSignal(props.override?.isTransparent ?? false);
	const [start, setStart] = createSignal(props.override?.startHour ?? "");
	const [pending, setPending] = createSignal(false);
	const [error, setError] = createSignal(false);
	onMount(() => dialog.showModal());

	async function save(reset = false) {
		if (pending()) return;
		setPending(true);
		setError(false);
		const updated = {
			classHash: props.classInfo.hash,
			isHidden: reset ? false : hidden(),
			isTransparent: reset ? false : transparent(),
			startHour: reset ? "" : start(),
		} as ClassOverride;
		try {
			if (!(await setClassOverride(updated))) {
				setError(true);
				return;
			}
			props.onSaved(updated);
			dialog.close();
		} catch {
			setError(true);
		} finally {
			setPending(false);
		}
	}

	return (
		<dialog
			ref={dialog}
			{...attrs(stylex.props(styles.dialog))}
			aria-labelledby={`${id}-title`}
			onClick={(event) => {
				if (event.target !== dialog || pending()) return;
				const bounds = dialog.getBoundingClientRect();
				if (event.clientX < bounds.left || event.clientX > bounds.right || event.clientY < bounds.top || event.clientY > bounds.bottom) {
					dialog.close();
				}
			}}
			onClose={props.onClose}
			onCancel={(event) => {
				if (pending()) event.preventDefault();
			}}
		>
			<form
				{...attrs(stylex.props(styles.form))}
				onSubmit={(event) => {
					event.preventDefault();
					void save();
				}}
				aria-busy={pending()}
			>
				<header {...attrs(stylex.props(styles.header))}>
					<Show when={props.classInfo.type}><span {...attrs(stylex.props(timetableStyles.eventType))}>{props.classInfo.type}</span></Show>
					<h2 id={`${id}-title`} {...attrs(stylex.props(styles.title, styles.headerTitle))}>
						{props.classInfo.name}
					</h2>
					<Show when={props.classInfo.teacher}>
						<p {...attrs(stylex.props(styles.meta))}>{props.classInfo.teacher}</p>
					</Show>
					<div {...attrs(stylex.props(styles.headerFooter))}>
						<span {...attrs(stylex.props(styles.meta))}>{props.classInfo.location}</span>
						<span {...attrs(stylex.props(styles.meta))}>{DayLabels[props.classInfo.day] ?? props.classInfo.day} · {props.classInfo.startTime}–{props.classInfo.endTime}</span>
					</div>
				</header>
				<div {...attrs(stylex.props(styles.body))}>
					<label {...attrs(stylex.props(styles.label))}>
						<span {...attrs(stylex.props(styles.copy))}>
							Ukryj w kalendarzu
						</span>
						<input
							{...attrs(stylex.props(styles.checkbox))}
							type="checkbox"
							checked={hidden()}
							disabled={pending()}
							onChange={(event) => setHidden(event.currentTarget.checked)}
						/>
					</label>
					<label {...attrs(stylex.props(styles.label))}>
						<span {...attrs(stylex.props(styles.copy))}>
							Zachowaj pierwotny termin
							<span {...attrs(stylex.props(styles.meta))}>
								Widoczny, ale nie blokuje czasu.
							</span>
						</span>
						<input
							{...attrs(stylex.props(styles.checkbox))}
							type="checkbox"
							checked={transparent()}
							disabled={pending()}
							onChange={(event) => setTransparent(event.currentTarget.checked)}
						/>
					</label>
					<label {...attrs(stylex.props(styles.label))}>
						<span {...attrs(stylex.props(styles.copy))}>
							Przesuń zajęcia na
							<span {...attrs(stylex.props(styles.meta))}>Puste = bez zmiany.</span>
						</span>
						<input
							{...attrs(stylex.props(styles.time))}
							type="time"
							value={start()}
							disabled={pending() || hidden()}
							onInput={(event) => setStart(event.currentTarget.value)}
						/>
					</label>
					<p {...attrs(stylex.props(styles.meta, styles.hint))}>
						Zmiany w kalendarzu po następnej synchronizacji.
					</p>
					<Show when={error()}>
						<p {...attrs(stylex.props(styles.error))} role="alert">
							Nie udało się zapisać zmian. Spróbuj ponownie.
						</p>
					</Show>
				</div>
				<div {...attrs(stylex.props(styles.actions))}>
					<button
						{...attrs(stylex.props(primitiveStyles.buttonBase, primitiveStyles.buttonSecondary, styles.reset))}
						type="button"
						disabled={pending()}
						onClick={() => void save(true)}
					>
						Przywróć domyślne
					</button>
					<button
						{...attrs(stylex.props(primitiveStyles.buttonBase, primitiveStyles.buttonSecondary))}
						type="button"
						disabled={pending()}
						onClick={() => dialog.close()}
					>
						Anuluj
					</button>
					<button
						{...attrs(stylex.props(primitiveStyles.buttonBase, primitiveStyles.buttonPrimary))}
						type="submit"
						disabled={pending()}
					>
						{pending() ? "Zapisywanie…" : "Zapisz"}
					</button>
				</div>
			</form>
		</dialog>
	);
}
