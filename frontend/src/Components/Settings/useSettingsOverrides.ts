import { type Accessor, createMemo, createSignal } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

type MutateOverrides = (mutator: (prev: Map<string, ClassOverride> | undefined) => Map<string, ClassOverride>) => void;

export function useSettingsOverrides(
	classes: Accessor<ClassInfo[] | undefined>,
	overrides: Accessor<Map<string, ClassOverride> | undefined>,
	mutateOverrides: MutateOverrides
) {
	const [addedClassHashes, setAddedClassHashes] = createSignal<Set<string>>(new Set());
	const [selectedClassHash, setSelectedClassHash] = createSignal("");

	function handleOverrideSaved(updated: ClassOverride) {
		mutateOverrides((prev) => {
			const next = new Map(prev ?? []);
			if (!updated.isHidden && !updated.isTransparent && updated.startHour === "") {
				next.delete(updated.classHash);
				setAddedClassHashes((old) => {
					const newSet = new Set(old);
					newSet.delete(updated.classHash);
					return newSet;
				});
			} else {
				next.set(updated.classHash, updated);
			}
			return next;
		});
	}

	const visibleClasses = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = addedClassHashes();
		return classesValue.filter((cls) => overridesValue?.has(cls.hash) || added.has(cls.hash));
	});

	const classesAvailableToAdd = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = addedClassHashes();
		return classesValue.filter((cls) => !overridesValue?.has(cls.hash) && !added.has(cls.hash));
	});

	function addSelectedClassToOverrides() {
		const classHash = selectedClassHash();
		if (!classHash) {
			return;
		}
		setAddedClassHashes((old) => {
			const newSet = new Set(old);
			newSet.add(classHash);
			return newSet;
		});
		setSelectedClassHash("");
	}

	return {
		handleOverrideSaved,
		visibleClasses,
		classesAvailableToAdd,
		addSelectedClassToOverrides,
		selectedClassHash,
		setSelectedClassHash,
	};
}
