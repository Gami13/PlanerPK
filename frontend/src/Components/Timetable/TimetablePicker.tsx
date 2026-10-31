import * as stylex from "@stylexjs/stylex";
import { Show, createUniqueId } from "solid-js";
import type { ColumnGroup } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs } from "../../utils";
import { ColumnOptions } from "../ColumnOptions";

const styles = stylex.create({
	row: { display: "flex", alignItems: "center", flexWrap: "wrap", gap: "8px 12px", paddingInline: "16px" },
	label: { fontSize: "0.85rem", color: "#576475" },
	compactRow: { paddingInline: 0 },
	compactSelect: { minWidth: "min(100%, 10rem)" },
	select: {
		padding: "8px 10px",
		borderWidth: "1px",
		borderStyle: "solid",
		borderColor: "#dfe5ed",
		borderRadius: "5px",
		backgroundColor: "#fff",
		color: "#0b2a45",
		fontSize: "0.85rem",
		maxWidth: "100%",
		minWidth: "min(100%, 16rem)",
	},
});

export function TimetablePicker(props: {
	columns: ColumnGroup[];
	selected?: number;
	compact?: boolean;
	onChange: (column: number) => void;
}) {
	const id = createUniqueId();
	return (
		<div {...attrs(stylex.props(styles.row, props.compact && styles.compactRow))}>
			<Show when={!props.compact}>
				<label for={id} {...attrs(stylex.props(styles.label))}>
					Wybierz plan zajęć
				</label>
			</Show>
			<select
				id={id}
				{...attrs(stylex.props(styles.select, props.compact && styles.compactSelect))}
				aria-label="Wybierz plan zajęć"
				value={props.selected ?? ""}
				onChange={(event) => props.onChange(Number(event.currentTarget.value))}
				disabled={!props.columns.length}
			>
				<ColumnOptions columns={props.columns} selected={props.selected} />
			</select>
		</div>
	);
}
