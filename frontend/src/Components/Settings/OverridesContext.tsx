import { type Accessor, type JSX, createContext } from "solid-js";
import type { ClassInfo, ClassOverride } from "../../gen/pkplaner/v1/pkplaner_pb";

type OverridesContextValue = {
	classes: Accessor<ClassInfo[] | undefined>;
	overrides: Accessor<Map<string, ClassOverride> | undefined>;
	selected: Accessor<string>;
	setSelected: (value: string) => void;
	addSelected: () => void;
	visibleClasses: Accessor<ClassInfo[]>;
	availableClasses: Accessor<ClassInfo[]>;
	handleOverrideSaved: (override: ClassOverride) => void;
};

export const OverridesContext = createContext<OverridesContextValue>();

type OverridesProviderProps = {
	value: OverridesContextValue;
	children: JSX.Element;
};

export function OverridesProvider(props: OverridesProviderProps) {
	return <OverridesContext.Provider value={props.value}>{props.children}</OverridesContext.Provider>;
}
