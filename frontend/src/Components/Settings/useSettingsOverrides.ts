import { type Accessor, createMemo, createSignal } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

type MutateOverrides = (mutator: (prev: Map<string, ClassOverride> | undefined) => Map<string, ClassOverride>) => void;

export function useSettingsOverrides(
	classes: Accessor<ClassInfo[] | undefined>,
	overrides: Accessor<Map<string, ClassOverride> | undefined>,
	mutateOverrides: MutateOverrides
) {
	const [selected, setSelected] = createSignal("");
	const [added, setAdded] = createSignal(new Set<string>());

	const visibleClasses = createMemo(() => {
		const list = classes() ?? [];
		const saved = overrides();
		const addedSet = added();
		return list.filter((c) => saved?.has(c.hash) || addedSet.has(c.hash));
	});

	const availableClasses = createMemo(() => {
		const list = classes() ?? [];
		const saved = overrides();
		const addedSet = added();
		return list.filter((c) => !saved?.has(c.hash) && !addedSet.has(c.hash));
	});

	const handleOverrideSaved = (u: ClassOverride) => {
		const hasNoOverrides = !u.isHidden && !u.isTransparent && !u.startHour;

		mutateOverrides((prev) => {
			const updated = new Map(prev ?? []);
			if (hasNoOverrides) {
				updated.delete(u.classHash);
				setAdded((added) => new Set([...added].filter((hash) => hash !== u.classHash)));
			} else {
				updated.set(u.classHash, u);
			}
			return updated;
		});
	};

	const addSelected = () => {
		const hash = selected();
		if (hash) {
			setAdded((h) => new Set([...h, hash]));
			setSelected("");
		}
	};

	return {
		selected,
		setSelected,
		addSelected,
		visibleClasses,
		availableClasses,
		handleOverrideSaved,
	};
}
