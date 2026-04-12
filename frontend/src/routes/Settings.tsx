import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, useContext } from "solid-js";

import { PageTitle } from "../Components/PageTitle";

import { AppContext } from "../appContext";
import { AccountStatusCard } from "../Components/Settings/AccountStatusCard";
import { ColumnSelectionCard } from "../Components/Settings/ColumnSelectionCard";
import { OverridesProvider } from "../Components/Settings/OverridesContext";
import { OverridesSection } from "../Components/Settings/OverridesSection";
import { SettingsIntro } from "../Components/Settings/SettingsIntro";
import { settingsStyles } from "../Components/Settings/settingsStyles";
import { useSettingsOverrides } from "../Components/Settings/useSettingsOverrides";
import { UIText } from "../constants";
import { getClasses, getColumns, getOverrides, setChosenColumn, setUserActive } from "../services/pkplaner";
import { attrs, getGroupByColumnIndex } from "../utils";

export function Settings() {
	const AppState = useContext(AppContext);
	const [columns] = createResource(getColumns);
	const [classes, { refetch: refetchClasses }] = createResource(getClasses);
	const [overrides, { mutate: mutateOverrides }] = createResource(getOverrides);

	const { selected, setSelected, addSelected, visibleClasses, availableClasses, handleOverrideSaved } =
		useSettingsOverrides(classes, overrides, mutateOverrides);

	const chosenColumn = createMemo(() => Number(AppState.store.userInfo?.chosenColumn ?? 0n));

	const currentColumnLabel = createMemo(() => {
		const selectedColumn = chosenColumn();
		if (selectedColumn === 0) {
			return UIText.settings.column.notSelected;
		}

		return getGroupByColumnIndex(columns(), selectedColumn)?.label ?? UIText.settings.column.loadingColumns;
	});

	async function toggleAccountStatus() {
		if (!AppState.store.userInfo) {
			return;
		}

		const newStatus = !AppState.store.userInfo.isActive;
		const success = await setUserActive(newStatus);
		if (!success) {
			alert(UIText.settings.account.statusError);
			return;
		}

		AppState.setStore("userInfo", "isActive", newStatus);
		alert(`Konto oznaczono jako ${newStatus ? "aktywne" : "nieaktywne"}.`);
	}

	async function changeChosenColumn(nextColumn: number) {
		const success = await setChosenColumn(nextColumn);
		if (!success) {
			alert(UIText.settings.column.errorSetting);
			return;
		}

		alert(UIText.settings.column.successSetting);
		AppState.setStore("userInfo", "chosenColumn", BigInt(nextColumn));
		refetchClasses();
	}

	return (
		<main {...attrs(stylex.props(settingsStyles.main))}>
			<PageTitle title="Ustawienia" />

			<section {...attrs(stylex.props(settingsStyles.basicSection))}>
				<SettingsIntro />

				<AccountStatusCard isActive={AppState.store.userInfo?.isActive} onToggle={toggleAccountStatus} />

				<ColumnSelectionCard
					chosenColumn={chosenColumn()}
					currentColumnLabel={currentColumnLabel()}
					columns={columns()}
					columnsLoading={columns.loading}
					columnsError={columns.error}
					onChangeChosenColumn={changeChosenColumn}
				/>

				<OverridesProvider
					value={{
						classes,
						overrides,
						selected,
						setSelected,
						addSelected,
						visibleClasses,
						availableClasses,
						handleOverrideSaved,
					}}
				>
					<OverridesSection
						chosenColumn={chosenColumn()}
						classesLoading={classes.loading}
						classesError={classes.error}
					/>
				</OverridesProvider>
			</section>
		</main>
	);
}
