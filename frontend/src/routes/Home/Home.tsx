import { A } from '@solidjs/router';
import { createResource, Show, useContext } from 'solid-js';
import { AppContext } from '../..';
import { UserCard } from '../../Components/UserCard/UserCard';
import { Endpoints, GOOGLE_CALENDAR_URL, PK_WEBSITE_URL, UserInfoResponse } from '../../constants';
import { setAccessToken } from '../../utils';
import styles from './home.module.css';

async function fetchAuthURL() {
  const response = await fetch(Endpoints.GetAuthURL);
  return response.text();
}


function logOut() {
  setAccessToken('');
  window.location.reload();
}

export function Home() {
  const AppState = useContext(AppContext);
  const [authURL] = createResource(fetchAuthURL);
  console.log(AppState.store.userInfo);
  return (
    <main class={styles.main}>

      <section class={styles.basic_section}>
        <div class={styles.description}>
          <h1 class={styles.header}>Google Calendar Timetable</h1>

          <p>Aplikacja pozwalająca na śledzenie rozkładów zajęć <a href={PK_WEBSITE_URL}>Politechniki Krakowskiej im. Tadeusza Kościuszki</a>. Po zalogowaniu się za pomocą konta Google, podaniu swojego wydziału, kierunku, a także grupy, aplikacja automatycznie będzie wprowadzać zmiany w <a href={GOOGLE_CALENDAR_URL}>kalendarz Google</a> prosto z rozkładów zajęć .xls zamieszczonych na stronie PK.</p>



        </div>
        <div class={styles.interactions_container}>
          <Show when={AppState.store.userInfo?.email} fallback={<a class={styles.register_button} href={authURL()}>Zarejestruj się z Google</a>}
          >
            <UserCard userInfo={AppState.store.userInfo as UserInfoResponse} />
            <div class={styles.auth_buttons}>
              <A href='/settings' class={styles.primary_button}>Ustawienia</A>
              <button class={styles.secondary_button} onClick={logOut}>Wyloguj</button>
            </div>

          </Show>
        </div>

        <div class={styles.image_container}>
          <img class={styles.google_calendar_preview} src="/google-calendar-preview.png" />
        </div>
      </section>
    </main>
  );
}
