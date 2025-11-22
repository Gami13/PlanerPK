/* @refresh reload */
import 'solid-devtools';
import './index.css';

import { render } from 'solid-js/web';

import { Route, Router, RouteSectionProps } from '@solidjs/router';
import { Component, createContext, onMount, Suspense } from 'solid-js';


import { createStore, SetStoreFunction } from 'solid-js/store';
import { Endpoints, UserInfoResponse } from './constants';
import { Authorize } from './routes/Authorize/Authorize';
import { Home } from './routes/Home/Home';
import { NotFound } from './routes/NotFound/NotFound';
import { Settings } from './routes/Settings/Settings';
import { getAccessToken } from './utils';

const root = document.getElementById('root');


if (!root) {
  throw new Error(
    'Root element not found. Did you forget to add it to your index.html? Or maybe the id attribute got misspelled?',
  );
}



export type AppStore = {
  userInfo: UserInfoResponse | undefined;
}

interface AppContextValue {
  store: AppStore;
  setStore: SetStoreFunction<{
    userInfo: UserInfoResponse | undefined;
  }>
}
const defaultContext: AppContextValue = {
  store: { userInfo: undefined },
  setStore: () => { },
};

export const AppContext = createContext<AppContextValue>(defaultContext);

async function fetchUserInfo(accessToken: string): Promise<UserInfoResponse | undefined> {
  const response = await fetch(Endpoints.GetUserInfo, {
    headers: {
      "Authorization": accessToken,

    },
  });

  if (response.status === 401) {
    return undefined;
  }
  return response.json();
}


export function App(props: { children: Element }) {


  const [store, setStore] = createStore({
    userInfo: {} as UserInfoResponse | undefined,
  });

  onMount(async () => {
    const userToken = getAccessToken();
    if (userToken) {
      const userInfo = await fetchUserInfo(userToken);
      setStore({ userInfo });
      console.log("Fetched user info on app mount:", userInfo);
    }
  });



  return (
    <AppContext.Provider value={{ store, setStore }}>
      <Suspense> {props.children}</Suspense >
    </AppContext.Provider>
  );
};



render(
  () => <Router root={App as unknown as Component<RouteSectionProps<unknown>>}>
    <Route path="/" component={Home} />
    <Route path="/authorize" component={Authorize} />
    <Route path="/settings" component={Settings} />
    <Route path="**" component={NotFound} />
  </Router>,
  root,
);
