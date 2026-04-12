import { rpc } from "../apiClient";
import type {
	AuthorizeResponse,
	ClassInfo,
	ClassOverride,
	ColumnGroup,
	GetUserInfoResponse,
} from "../gen/pkplaner/v1/pkplaner_pb";

const dayOrder = ["Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"];

const getDayOrderIndex = (day: string) => {
	const index = dayOrder.indexOf(day);
	return index === -1 ? dayOrder.length : index;
};

function sortClassesBySchedule(classes: ClassInfo[]): ClassInfo[] {
	return [...classes].sort((a, b) => {
		const dayDiff = getDayOrderIndex(a.day) - getDayOrderIndex(b.day);
		if (dayDiff !== 0) {
			return dayDiff;
		}
		return a.startTime.localeCompare(b.startTime);
	});
}

async function withFallback<T>(request: () => Promise<T>, fallback: T): Promise<T> {
	try {
		return await request();
	} catch {
		return fallback;
	}
}

async function runMutation(request: () => Promise<unknown>): Promise<boolean> {
	try {
		await request();
		return true;
	} catch {
		return false;
	}
}

export async function getAuthURL() {
	return (await rpc.getAuthURL({})).url;
}

export async function authorizeByCode(code: string): Promise<AuthorizeResponse> {
	return rpc.authorize({ code });
}

export async function getUserInfoSafe(): Promise<GetUserInfoResponse | undefined> {
	return withFallback(() => rpc.getUserInfo({}), undefined);
}

export async function setChosenColumn(chosenColumn: number): Promise<boolean> {
	return runMutation(() => rpc.setChosenColumn({ chosenColumn }));
}

export async function setUserActive(isActive: boolean): Promise<boolean> {
	return runMutation(() => rpc.setUserActive({ isActive }));
}

export async function getColumns(): Promise<ColumnGroup[]> {
	return (await rpc.getColumns({})).columns;
}

export async function getClasses(): Promise<ClassInfo[]> {
	const classes = await withFallback(async () => (await rpc.getClasses({})).classes, [] as ClassInfo[]);
	return sortClassesBySchedule(classes);
}

export async function getOverrides(): Promise<Map<string, ClassOverride>> {
	const overrides = await withFallback(async () => (await rpc.getOverrides({})).overrides, [] as ClassOverride[]);
	return new Map(overrides.map((override) => [override.classHash, override] as const));
}

export async function setClassOverride(override: ClassOverride): Promise<boolean> {
	return runMutation(() => rpc.setOverride({ override }));
}
