import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, createSignal, Show, useContext } from "solid-js";

import { PageTitle } from "../Components/PageTitle";
import { TimetablePreview } from "../Components/Timetable/TimetablePreview";

import { AppContext } from "../appContext";
import { AccountStatusToggle } from "../Components/Settings/AccountStatusToggle";
import { ColumnSelectionForm } from "../Components/Settings/ColumnSelectionForm";
import { SettingCard } from "../Components/Settings/SettingCard";
import { ClassOverrideDialog } from "../Components/Settings/ClassOverrideDialog";
import type { ClassInfo, ClassOverride } from "../gen/pkplaner/v1/pkplaner_pb";
import { SettingsIntro } from "../Components/Settings/SettingsIntro";
import { UIText } from "../constants";
import { getClasses, getColumns, getOverrides, setChosenColumn, setUserActive } from "../services/pkplaner";
import { spacing } from "../styles/tokens.stylex";
import { attrs, getGroupByColumnIndex } from "../utils";

const style = stylex.create({
	main: {
		padding: spacing.xl,
		gap: spacing.xl,
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	basicSection: {
		gap: spacing.md,
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
});
export function Settings() {
	const AppState = useContext(AppContext);
	const [columns] = createResource(getColumns);
	const [classes, { refetch: refetchClasses }] = createResource(getClasses);
	const [overrides, { mutate: mutateOverrides }] = createResource(getOverrides);
	const loadedClasses = () => (classes.error ? undefined : classes());
	const loadedOverrides = () => (overrides.error ? undefined : overrides());

	const [editingClass, setEditingClass] = createSignal<ClassInfo>();
	const [showHidden, setShowHidden] = createSignal(false);
	function handleOverrideSaved(updated: ClassOverride) {
		mutateOverrides((previous) => {
			const next = new Map(previous ?? []);
			if (!updated.isHidden && !updated.isTransparent && !updated.startHour) next.delete(updated.classHash);
			else next.set(updated.classHash, updated);
			return next;
		});
	}

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
		<main {...attrs(stylex.props(style.main))}>
			<PageTitle title="Ustawienia" />

			<section {...attrs(stylex.props(style.basicSection))}>
				<SettingsIntro />

				<SettingCard
					title={UIText.settings.account.title}
					headerAction={
						<AccountStatusToggle isActive={AppState.store.userInfo?.isActive} onToggle={toggleAccountStatus} />
					}
				>
					<ColumnSelectionForm
						chosenColumn={chosenColumn()}
						currentColumnLabel={currentColumnLabel()}
						columns={columns()}
						columnsLoading={columns.loading}
						columnsError={columns.error}
						onChangeChosenColumn={changeChosenColumn}
					/>
				</SettingCard>

				<TimetablePreview
					onClassSelect={setEditingClass}
					showHidden={showHidden()}
					onShowHiddenChange={setShowHidden}
					classes={loadedClasses()}
					overrides={loadedOverrides()}
					loading={classes.loading || overrides.loading}
					error={classes.error || overrides.error}
					groupLabel={currentColumnLabel()}
					noGroup={chosenColumn() === 0}
				/>

				<Show when={editingClass()} keyed>
					{(classInfo) => (
						<ClassOverrideDialog
							classInfo={classInfo}
							override={loadedOverrides()?.get(classInfo.hash)}
							onSaved={handleOverrideSaved}
							onClose={() => setEditingClass(undefined)}
						/>
					)}
				</Show>
			</section>
		</main>
	);
}
