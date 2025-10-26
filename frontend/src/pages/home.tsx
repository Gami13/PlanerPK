import { createSignal } from 'solid-js';
import styles from './home.module.css'

export default function Home() {
    const registerWithGoogle = (event) => {
        // TODO: Register with google
        console.warn("TODO: Register with google");
    };
  return (
    <main class={styles.main}>
      <section class={styles.basic_section}>
        <div>
          <h1 class={styles.header}>Google Calendar Timetable</h1>

          <p>Aplikacja pozwalająca na śledzenie rozkładów zajęć <a href="https://www.pk.edu.pl/index.php?lang=pl">Politechniki Krakowskiej im. Tadeusza Kościuszki</a>. Po zalogowaniu się za pomocą konta Google, podaniu swojego wydziału, kierunku, a także grupy, aplikacja automatycznie będzie wprowadzać zmiany w <a href="https://calendar.google.com/calendar?hl=pl">kalendarz Google</a> prosto z rozkładów zajęć .xls zamieszczonych na stronie PK.</p>

          <button class={styles.register_button} onClick={[registerWithGoogle]}>Zarejestruj się z Google</button>
        </div>

        <div>
          <img class={styles.google_calendar_preview} src="/google-calendar-preview.png"/>
        </div>
      </section>
    </main>
  );
}
