/* @refresh reload */
import "./index.css";

import { render } from "solid-js/web";

import { Route, Router, RouteSectionProps } from "@solidjs/router";
import { Component, onMount, ParentProps, Suspense } from "solid-js";

import { MetaProvider } from "@solidjs/meta";
import { createStore } from "solid-js/store";
import { AppContext, type AppStore, updateLoggedInUser } from "./appContext";
import { Authorize } from "./routes/Authorize";
import { Home } from "./routes/Home";
import { NotFound } from "./routes/NotFound";
import { Settings } from "./routes/Settings";

if (import.meta.env.DEV) {
	void import("solid-devtools");
}

const root = document.getElementById("root");

if (!root) {
	throw new Error(
		"Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?"
	);
}

export function App(props: ParentProps) {
	const [store, setStore] = createStore<AppStore>({ userInfo: undefined });

	onMount(() => {
		void updateLoggedInUser(setStore);
	});

	return (
		<AppContext.Provider value={{ store, setStore }}>
			<Suspense>{props.children}</Suspense>
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
