import * as stylex from "@stylexjs/stylex";
import { For, Show, createUniqueId } from "solid-js";
import type { ClassInfo } from "../../gen/pkplaner/v1/pkplaner_pb";
import { borderRadius, colors, spacing } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";
import { defaultTitleTemplate, formatClassTitle, titlePlaceholders } from "./titleTemplate";

const styles = stylex.create({
	editor: { display: "flex", flexDirection: "column", gap: spacing.sm, minWidth: 0 },
	input: {
		width: "100%",
		padding: "10px 12px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.md,
		fontSize: "0.9em",
		color: colors.primaryText,
		backgroundColor: colors.white,
	},
	hint: { margin: 0, color: colors.secondaryText, fontSize: "0.82em" },
	fields: { display: "flex", flexWrap: "wrap", gap: spacing.sm },
	field: { fontSize: "0.8em", color: colors.secondaryText },
	output: { display: "block", overflowWrap: "anywhere", fontSize: "0.9em", color: colors.primaryText },
	reset: {
		alignSelf: "flex-start",
		backgroundColor: colors.white,
		color: colors.primary,
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: colors.defaultBorder,
		borderRadius: borderRadius.sm,
		padding: "6px 10px",
		cursor: "pointer",
		fontSize: "0.82em",
	},
});

export function TitleTemplateEditor(props: {
	value: string;
	onChange: (value: string) => void;
	sampleClass?: ClassInfo;
}) {
	const id = createUniqueId();
	return (
		<div {...attrs(stylex.props(styles.editor))}>
			<label for={id}>Szablon tytułu</label>
			<input
				{...attrs(stylex.props(styles.input))}
				id={id}
				type="text"
				value={props.value}
				onInput={(event) => props.onChange(event.currentTarget.value)}
				aria-describedby={`${id}-hint`}
			/>
			<p id={`${id}-hint`} {...attrs(stylex.props(styles.hint))}>
				Zmiany są widoczne od razu w podglądzie poniżej. Nie są zapisywane ani wysyłane do Kalendarza Google.
			</p>
			<div {...attrs(stylex.props(styles.fields))}>
				<For each={titlePlaceholders}>
					{(field) => <code {...attrs(stylex.props(styles.field))}>{`{${field}}`}</code>}
				</For>
			</div>
			<Show when={props.sampleClass}>
				{(sample) => (
					<output for={id} {...attrs(stylex.props(styles.output))} aria-live="polite">
						{formatClassTitle(props.value, sample())}
					</output>
				)}
			</Show>
			<button {...attrs(stylex.props(styles.reset))} type="button" onClick={() => props.onChange(defaultTitleTemplate)}>
				Przywróć domyślny szablon
			</button>
		</div>
	);
}
