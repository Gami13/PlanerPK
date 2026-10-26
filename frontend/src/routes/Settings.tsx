import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, createSignal, Show, useContext } from "solid-js";

import { PageTitle } from "../Components/PageTitle";
import { pageStyles } from "../styles/pageStyles";
import { TimetablePreview } from "../Components/Timetable/TimetablePreview";

import { AppContext } from "../appContext";
import { Button } from "../Components/Button";
import { AccountStatusToggle } from "../Components/Settings/AccountStatusToggle";
import { ColumnSelectionForm } from "../Components/Settings/ColumnSelectionForm";
import { ClassOverrideDialog } from "../Components/Settings/ClassOverrideDialog";
import { AddClassDialog } from "../Components/Settings/AddClassDialog";
import type { ClassInfo, ClassOverride } from "../gen/pkplaner/v1/pkplaner_pb";
import { SettingsIntro } from "../Components/Settings/SettingsIntro";
import { UIText } from "../constants";
import { getClasses, getColumns, getOverrides, setChosenColumn, setUserActive } from "../services/pkplaner";
import { attrs, getGroupByColumnIndex } from "../utils";

const styles = stylex.create({
	controls: { display: "flex", alignItems: "center", flexWrap: "wrap", gap: "12px 20px", minWidth: 0 },
	sync: { display: "flex", alignItems: "center", gap: "8px", color: "#576475", fontSize: "0.82rem" },
});

export function Settings() {
	const AppState = useContext(AppContext);
	const [columns] = createResource(getColumns);
	const [classes, { refetch: refetchClasses }] = createResource(getClasses);
	const [overrides, { mutate: mutateOverrides }] = createResource(getOverrides);
	const loadedClasses = () => (classes.error ? undefined : classes.latest);
	const loadedOverrides = () => (overrides.error ? undefined : overrides.latest);

	const [editingClass, setEditingClass] = createSignal<ClassInfo>();
	const [showHidden, setShowHidden] = createSignal(false);
	const [addingClass, setAddingClass] = createSignal(false);
	function handleOverrideSaved(updated: ClassOverride) {
		mutateOverrides((previous) => {
			const next = new Map(previous ?? []);
			if (!updated.isHidden && !updated.isTransparent && !updated.startHour && !updated.sourceColumn)
				next.delete(updated.classHash);
			else next.set(updated.classHash, updated);
			return next;
		});
		void refetchClasses();
	}

	const chosenColumn = createMemo(() => Number(AppState.store.userInfo?.chosenColumn ?? 0n));

	const currentColumnLabel = createMemo(() => {
		const selectedColumn = chosenColumn();
		if (selectedColumn === 0) {
			return UIText.settings.column.notSelected;
		}

		return getGroupByColumnIndex(columns(), selectedColumn)?.label ?? UIText.settings.column.loadingColumns;
	});
	const previewGroupLabel = createMemo(() => {
		const group = getGroupByColumnIndex(columns(), chosenColumn());
		return group ? `${group.category} · ${group.label}` : currentColumnLabel();
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
		<main {...attrs(stylex.props(pageStyles.page))}>
			<PageTitle title="Ustawienia" />

			<section {...attrs(stylex.props(pageStyles.section))}>
				<SettingsIntro>
					<div {...attrs(stylex.props(styles.controls))}>
						<div {...attrs(stylex.props(styles.sync))}>
							<span>{UIText.settings.account.title}</span>
							<AccountStatusToggle isActive={AppState.store.userInfo?.isActive} onToggle={toggleAccountStatus} />
						</div>
						<ColumnSelectionForm
							compact
							chosenColumn={chosenColumn()}
							columns={columns()}
							columnsLoading={columns.loading}
							columnsError={columns.error}
							onChangeChosenColumn={changeChosenColumn}
						/>
					</div>
				</SettingsIntro>

				<TimetablePreview
					headerAction={
						<Button
							type="button"
							disabled={!chosenColumn() || classes.loading || !!classes.error || overrides.loading || !!overrides.error}
							onClick={() => setAddingClass(true)}
						>
							+ Dodaj zajęcia
						</Button>
					}
					expand
					onClassSelect={setEditingClass}
					showHidden={showHidden()}
					onShowHiddenChange={setShowHidden}
					classes={loadedClasses()}
					overrides={loadedOverrides()}
					loading={classes.loading || overrides.loading}
					error={classes.error || overrides.error}
					groupLabel={previewGroupLabel()}
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
				<Show when={addingClass()}>
					<AddClassDialog
						existingClasses={loadedClasses() ?? []}
						onAdded={handleOverrideSaved}
						onClose={() => setAddingClass(false)}
					/>
				</Show>
			</section>
		</main>
	);
}
