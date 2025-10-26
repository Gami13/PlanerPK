import { createSignal, createResource } from 'solid-js';
import styles from './home.module.css'

const fetchAuthURL = async () => {
  const response = await fetch(`http://localhost:8080/api/getAuthURL`);
  return response.text();
}

const registerWithGoogle = (authURL, event) => {
  window.open(authURL(), "_self");
};

export default function Home() {
  const [authURL] = createResource(fetchAuthURL);

  return (
    <main class={styles.main}>
      <section class={styles.basic_section}>
        <div>
          <h1 class={styles.header}>Google Calendar Timetable</h1>

          <p>Aplikacja pozwalająca na śledzenie rozkładów zajęć <a href="https://www.pk.edu.pl/index.php?lang=pl">Politechniki Krakowskiej im. Tadeusza Kościuszki</a>. Po zalogowaniu się za pomocą konta Google, podaniu swojego wydziału, kierunku, a także grupy, aplikacja automatycznie będzie wprowadzać zmiany w <a href="https://calendar.google.com/calendar?hl=pl">kalendarz Google</a> prosto z rozkładów zajęć .xls zamieszczonych na stronie PK.</p>

          <button class={styles.register_button} onClick={[registerWithGoogle, authURL]}>Zarejestruj się z Google</button>
        </div>

        <div>
          <img class={styles.google_calendar_preview} src="/google-calendar-preview.png"/>
        </div>
      </section>
    </main>
  );
}
