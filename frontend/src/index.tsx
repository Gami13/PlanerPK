/* @refresh reload */
import "solid-devtools";
import "./index.css";

import { render } from "solid-js/web";

import { Route, Router, RouteSectionProps } from "@solidjs/router";
import { Component, createContext, onMount, Suspense } from "solid-js";

import { MetaProvider } from "@solidjs/meta";
import { createStore, SetStoreFunction } from "solid-js/store";
import { rpc } from "./api_client";
import { ENV } from "./env";
import type { GetUserInfoResponse } from "./gen/pkplaner/v1/pkplaner_pb.js";
import { Authorize } from "./routes/Authorize";
import { Home } from "./routes/Home";
import { NotFound } from "./routes/NotFound";
import { Settings } from "./routes/Settings";
import { getAccessToken } from "./utils";

const root = document.getElementById("root");

if (!root) {
	throw new Error(
		"Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?"
	);
}

export type AppStore = {
	userInfo: GetUserInfoResponse | undefined;
};

interface AppContextValue {
	store: AppStore;
	setStore: SetStoreFunction<{
		userInfo: GetUserInfoResponse | undefined;
	}>;
}
const defaultContext: AppContextValue = {
	store: { userInfo: undefined },
	setStore: () => {},
};

export const AppContext = createContext<AppContextValue>(defaultContext);

async function fetchUserInfo(): Promise<GetUserInfoResponse | undefined> {
	try {
		return await rpc.getUserInfo({});
	} catch {
		return undefined;
	}
}

export async function updateLoggedInUser(store: AppStore, setStore: SetStoreFunction<AppStore>) {
	const userToken = getAccessToken();
	if (userToken) {
		fetchUserInfo().then((userInfo) => {
			setStore({ userInfo });
			console.log("Fetched user info on update:", userInfo);
		});
	}
}

export function App(props: { children: Element }) {
	const [store, setStore] = createStore({
		userInfo: {} as GetUserInfoResponse | undefined,
	});

	onMount(async () => {
		console.log(ENV);
		updateLoggedInUser(store, setStore);
	});

	return (
		<AppContext.Provider value={{ store, setStore }}>
			<Suspense> {props.children}</Suspense>
		</AppContext.Provider>
	);
}

render(
	() => (
		<MetaProvider>
			<Router root={App as unknown as Component<RouteSectionProps<unknown>>}>
				<Route path="/" component={Home} />
				<Route path="/authorize" component={Authorize} />
				<Route path="/settings" component={Settings} />
				<Route path="**" component={NotFound} />
			</Router>
		</MetaProvider>
	),
	root
);
