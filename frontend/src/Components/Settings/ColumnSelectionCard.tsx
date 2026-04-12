import * as stylex from "@stylexjs/stylex";
import { For, Show } from "solid-js";
import { Button } from "../../Components/Button";
import { UIText } from "../../constants";
import { ColumnGroup } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs, groupByCategory } from "../../utils";
import { SettingCard } from "./SettingCard";
import { settingsStyles } from "./settingsStyles";

type ColumnSelectionCardProps = {
	chosenColumn: number;
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
	},
	selectInput: {
		flex: 1,
	},
	submitButton: {
		marginLeft: "auto",
	},
});

export function ColumnSelectionCard(props: ColumnSelectionCardProps) {
	return (
		<SettingCard title={UIText.settings.column.title} description={UIText.settings.column.description}>
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
					<div {...attrs(stylex.props(columnFormStyles.selectRow))}>
						<select
							{...attrs(stylex.props(settingsStyles.columnSelect, columnFormStyles.selectInput))}
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
		</SettingCard>
	);
}
