import * as stylex from "@stylexjs/stylex";
import { createSignal, createUniqueId, onMount, Show } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";
import { setClassOverride } from "../../services/pkplaner";
import { borderRadius, colors, spacing } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";
import { DayLabels } from "../../constants";

const styles = stylex.create({
	dialog: {
		width: "min( calc(100% - 2rem), 30rem)",
		maxHeight: "calc(100dvh - 2rem)",
		overflowY: "auto",
		padding: spacing.lg,
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.lg,
		backgroundColor: colors.white,
		color: colors.primaryText,
		boxShadow: "0 16px 64px rgba(15, 23, 42, 0.2)",
		"::backdrop": { backgroundColor: "rgba(15, 23, 42, 0.4)" },
	},
	form: { display: "flex", flexDirection: "column", gap: spacing.md },
	title: { margin: 0, fontSize: "1.2em", fontWeight: 600 },
	meta: { margin: 0, fontSize: "0.85em", color: colors.secondaryText },
	label: { display: "flex", alignItems: "center", gap: spacing.sm, fontSize: "0.9em" },
	time: {
		padding: "8px 10px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.sm,
		fontSize: "1em",
		color: colors.primaryText,
	},
	actions: { display: "flex", flexWrap: "wrap", justifyContent: "flex-end", gap: spacing.sm },
	button: {
		padding: "8px 12px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.sm,
		backgroundColor: colors.white,
		color: colors.primaryText,
		fontSize: "0.9em",
		cursor: "pointer",
	},
	primary: { backgroundColor: "#e4eefb", borderColor: "#c9dcf5", color: "#245a85" },
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
				<h2 id={`${id}-title`} {...attrs(stylex.props(styles.title))}>
					{props.classInfo.name}
				</h2>
				<p {...attrs(stylex.props(styles.meta))}>
					{DayLabels[props.classInfo.day] ?? props.classInfo.day} · {props.classInfo.startTime}–
					{props.classInfo.endTime} · {props.classInfo.teacher}
				</p>
				<label {...attrs(stylex.props(styles.label))}>
					<input
						type="checkbox"
						checked={hidden()}
						disabled={pending()}
						onChange={(event) => setHidden(event.currentTarget.checked)}
					/>
					Ukryj zajęcia
				</label>
				<label {...attrs(stylex.props(styles.label))}>
					<input
						type="checkbox"
						checked={transparent()}
						disabled={pending()}
						onChange={(event) => setTransparent(event.currentTarget.checked)}
					/>
					Zostaw oryginalną pozycję jako transparentną
				</label>
				<label {...attrs(stylex.props(styles.label))}>
					Nowa godzina
					<input
						{...attrs(stylex.props(styles.time))}
						type="time"
						value={start()}
						disabled={pending() || hidden()}
						onInput={(event) => setStart(event.currentTarget.value)}
					/>
				</label>
				<p {...attrs(stylex.props(styles.meta))}>
					Pusta godzina zachowuje oryginalny termin. Zapisane zmiany trafią do Kalendarza Google przy następnej
					synchronizacji.
				</p>
				<Show when={error()}>
					<p {...attrs(stylex.props(styles.error))} role="alert">
						Nie udało się zapisać zmian. Spróbuj ponownie.
					</p>
				</Show>
				<div {...attrs(stylex.props(styles.actions))}>
					<button
						{...attrs(stylex.props(styles.button))}
						type="button"
						disabled={pending()}
						onClick={() => void save(true)}
					>
						Przywróć oryginał
					</button>
					<button
						{...attrs(stylex.props(styles.button))}
						type="button"
						disabled={pending()}
						onClick={() => dialog.close()}
					>
						Anuluj
					</button>
					<button {...attrs(stylex.props(styles.button, styles.primary))} type="submit" disabled={pending()}>
						{pending() ? "Zapisywanie…" : "Zapisz"}
					</button>
				</div>
			</form>
		</dialog>
	);
}
