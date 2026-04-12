import { createContext } from "solid-js";
import type { SetStoreFunction } from "solid-js/store";
import { GetUserInfoResponse } from "./gen/pkplaner/v1/pkplaner_pb";
import { getUserInfoSafe } from "./services/pkplaner";
import { getAccessToken } from "./utils";

export type AppStore = {
	userInfo: GetUserInfoResponse | undefined;
};

export type AppContextValue = {
	store: AppStore;
	setStore: SetStoreFunction<AppStore>;
};

const defaultContext: AppContextValue = {
	store: { userInfo: undefined },
	setStore: () => {},
};

export const AppContext = createContext<AppContextValue>(defaultContext);

export async function updateLoggedInUser(setStore: SetStoreFunction<AppStore>) {
	const userToken = getAccessToken();
	if (!userToken) {
		setStore("userInfo", undefined);
		return;
	}
	const userInfo = await getUserInfoSafe();
	setStore("userInfo", userInfo);
}
