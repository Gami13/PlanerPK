import { For } from "solid-js";
import type { ColumnGroup } from "../gen/pkplaner/v1/pkplaner_pb";
import { groupByCategory } from "../utils";

export function ColumnOptions(props: { columns: ColumnGroup[]; selected?: number }) {
	return (
		<For each={[...groupByCategory(props.columns)]}>
			{([category, columns]) => (
				<optgroup label={category || "Inne"}>
					<For each={columns}>
						{(column) => (
							<option value={column.column} selected={column.column === props.selected}>
								{column.label}
							</option>
						)}
					</For>
				</optgroup>
			)}
		</For>
	);
}
