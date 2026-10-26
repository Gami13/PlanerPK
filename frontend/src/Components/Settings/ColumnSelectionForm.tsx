import * as stylex from "@stylexjs/stylex";
import { For, Show } from "solid-js";
import { Button } from "../../Components/Button";
import { UIText } from "../../constants";
import { ColumnGroup } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs, groupByCategory } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type ColumnSelectionFormProps = {
	chosenColumn: number;
	compact?: boolean;
	currentColumnLabel: string;
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
						{...attrs(
							stylex.props(
								settingsStyles.columnSelect,
								columnFormStyles.selectInput,
								props.compact && columnFormStyles.compactSelect
							)
						)}
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
						<Show when={props.columns}>
							<For each={[...groupByCategory(props.columns!)]}>
								{([category, options]) => (
									<optgroup label={category || "Inne"}>
										<For each={options}>
											{(option) => (
												<option value={option.column} selected={props.chosenColumn === option.column}>
													{option.label}
												</option>
											)}
										</For>
									</optgroup>
								)}
							</For>
						</Show>
					</select>
					<div {...attrs(stylex.props(columnFormStyles.submitButton))}>
						<Button isPrimary={true}>{UIText.settings.column.change}</Button>
					</div>
				</div>
			</div>
		</form>
	);
}
