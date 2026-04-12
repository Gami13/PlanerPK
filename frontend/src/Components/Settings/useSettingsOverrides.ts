import { type Accessor, createMemo, createSignal } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

type MutateOverrides = (mutator: (prev: Map<string, ClassOverride> | undefined) => Map<string, ClassOverride>) => void;

type HookState = {
	addedClassHashes: Set<string>;
	selectedClassHash: string;
};

export function useSettingsOverrides(
	classes: Accessor<ClassInfo[] | undefined>,
	overrides: Accessor<Map<string, ClassOverride> | undefined>,
	mutateOverrides: MutateOverrides
) {
	const [state, setState] = createSignal<HookState>({
		addedClassHashes: new Set(),
		selectedClassHash: "",
	});

	function handleOverrideSaved(updated: ClassOverride) {
		mutateOverrides((prev) => {
			const next = new Map(prev ?? []);
			if (!updated.isHidden && !updated.isTransparent && updated.startHour === "") {
				next.delete(updated.classHash);
				setState((prev) => ({
					...prev,
					addedClassHashes: new Set([...prev.addedClassHashes].filter((h) => h !== updated.classHash)),
				}));
			} else {
				next.set(updated.classHash, updated);
			}
			return next;
		});
	}

	const visibleClasses = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = state().addedClassHashes;
		return classesValue.filter((cls) => overridesValue?.has(cls.hash) || added.has(cls.hash));
	});

	const classesAvailableToAdd = createMemo(() => {
		const classesValue = classes() ?? [];
		const overridesValue = overrides();
		const added = state().addedClassHashes;
		return classesValue.filter((cls) => !overridesValue?.has(cls.hash) && !added.has(cls.hash));
	});

	function addSelectedClassToOverrides() {
		const currentState = state();
		const classHash = currentState.selectedClassHash;
		if (!classHash) {
			return;
		}
		setState((prev) => ({
			selectedClassHash: "",
			addedClassHashes: new Set([...prev.addedClassHashes, classHash]),
		}));
	}

	return {
		handleOverrideSaved,
		visibleClasses,
		classesAvailableToAdd,
		addSelectedClassToOverrides,
		selectedClassHash: () => state().selectedClassHash,
		setSelectedClassHash: (value: string) => setState((prev) => ({ ...prev, selectedClassHash: value })),
	};
}
