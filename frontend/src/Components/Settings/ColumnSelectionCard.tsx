import * as stylex from "@stylexjs/stylex";
import { For, Show } from "solid-js";
import { Button } from "../../Components/Button";
import { ColumnGroup } from "../../gen/pkplaner/v1/pkplaner_pb";
import { attrs, groupByCategory } from "../../utils";
import { settingsStyles } from "./settingsStyles";

type ColumnSelectionCardProps = {
	chosenColumn: number;
	currentColumnLabel: string;
	columns: ColumnGroup[] | undefined;
	columnsLoading: boolean;
	columnsError: unknown;
	onChangeChosenColumn: (nextColumn: number) => Promise<void>;
};

export function ColumnSelectionCard(props: ColumnSelectionCardProps) {
	return (
		<form
			{...attrs(stylex.props(settingsStyles.setting))}
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
			<h2 {...attrs(stylex.props(settingsStyles.settingHeader))}>Wybrana kolumna</h2>
			<p>Zmianę kolumny wykonuj wyłącznie w uzasadnionych przypadkach, aby uniknąć ograniczeń konta.</p>
			<p {...attrs(stylex.props(settingsStyles.settingDescription))}>Obecna kolumna: {props.currentColumnLabel}</p>
			<select {...attrs(stylex.props(settingsStyles.columnSelect))} name="chosen_column" id="chosen_column">
				<Show when={props.columnsError}>
					<option disabled>Błąd ładowania kolumn</option>
				</Show>
				<Show when={props.columnsLoading}>
					<option disabled>Ładowanie...</option>
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
			<Button isPrimary={true}>Zmień kolumnę</Button>
		</form>
	);
}
