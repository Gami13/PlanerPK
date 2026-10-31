import * as stylex from "@stylexjs/stylex";
import { Show } from "solid-js";
import { Button } from "../../Components/Button";
import { ColumnOptions } from "../ColumnOptions";
import { UIText } from "../../constants";
import type { ColumnGroup } from "../../gen/pkplaner/v1/pkplaner_pb";
import { colors } from "../../styles/tokens.stylex";
import { attrs } from "../../utils";

type ColumnSelectionFormProps = {
	chosenColumn: number;
	compact?: boolean;
	columns: ColumnGroup[] | undefined;
	columnsLoading: boolean;
	columnsError: unknown;
	onChangeChosenColumn: (nextColumn: number) => Promise<void>;
};

const columnFormStyles = stylex.create({
	formContent: {
		display: "flex",
		flexDirection: "column",
		gap: "0.75em",
	},
	selectRow: {
		display: "flex",
		gap: "0.75em",
		alignItems: "center",
		width: "100%",
		flexWrap: "wrap",
	},
	selectInput: {
		borderColor: colors.defaultBorder,
		borderStyle: "solid",
		borderWidth: "1px",
		borderRadius: "5px",
		paddingBlock: "8px",
		paddingInline: "10px",
		color: colors.primaryText,
		fontSize: "0.85rem",
		backgroundColor: colors.white,
		maxWidth: "20em",
		width: "100%",
		outline: { default: "none", ":focus": "2px solid #039be5" },
		outlineOffset: { ":focus": "-1px" },
		flex: 1,
		minWidth: "min(100%, 12rem)",
	},
	compactSelect: { width: "9rem", minWidth: "8rem", maxWidth: "14rem" },
	label: {
		fontWeight: 500,
		fontSize: "0.82rem",
		color: "#576475",
	},
	submitButton: {
		flexShrink: 0,
	},
});

export function ColumnSelectionForm(props: ColumnSelectionFormProps) {
	return (
		<form
			method="post"
			onSubmit={async (e) => {
				e.preventDefault();
				const formData = new FormData(e.currentTarget);
				const chosenColumn = formData.get("chosen_column");
				if (chosenColumn) {
					await props.onChangeChosenColumn(Number(chosenColumn));
				}
			}}
		>
			<div {...attrs(stylex.props(columnFormStyles.formContent))}>
				<Show when={!props.compact}>
					<label for="chosen_column" {...attrs(stylex.props(columnFormStyles.label))}>
						{UIText.settings.column.title}
					</label>
				</Show>
				<div {...attrs(stylex.props(columnFormStyles.selectRow))}>
					<select
						{...attrs(stylex.props(columnFormStyles.selectInput, props.compact && columnFormStyles.compactSelect))}
						aria-label={UIText.settings.column.title}
						name="chosen_column"
						id="chosen_column"
					>
						<Show when={props.columnsError}>
							<option disabled>{UIText.settings.column.errorSetting}</option>
						</Show>
						<Show when={props.columnsLoading}>
							<option disabled>{UIText.settings.column.loadingColumns}</option>
						</Show>
						<ColumnOptions columns={props.columns ?? []} selected={props.chosenColumn} />
					</select>
					<div {...attrs(stylex.props(columnFormStyles.submitButton))}>
						<Button isPrimary={true}>{UIText.settings.column.change}</Button>
					</div>
				</div>
			</div>
		</form>
	);
}
