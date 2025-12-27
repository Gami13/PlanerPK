import { For, useContext } from 'solid-js';
import { AppContext } from '../..';
import { PageTitle } from '../../Components/PageTitle';
import { Endpoints } from '../../constants';
import { getAccessToken } from '../../utils';
import styles from './Settings.module.css';

async function changeChosenColumn(newColumn: number) {
  const response = await fetch(Endpoints.SetChosenColumn, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': getAccessToken() || '',
    },
    body: JSON.stringify({ chosenColumn: newColumn }),
  });
  return response.ok;
}

const columnOptions = [{ value: 2, label: 'Kolumna 2 (GK1)' },
{ value: 3, label: 'Kolumna 3 (GK2)' },
{ value: 4, label: 'Kolumna 4 (GK3)' },
{ value: 5, label: 'Kolumna 5 (GK4)' },
{ value: 6, label: 'Kolumna 6 (GK5)' },
{ value: 7, label: 'Kolumna 7 (GK6)' },
{ value: 8, label: 'Kolumna 8 (GK7)' }];

export function Settings() {
  const AppState = useContext(AppContext);

  return (
    <main class={styles.main}>
      <PageTitle title="Ustawienia" />

      <section class={styles.basic_section}>
        <div class={styles.description}>
          <h1 class={styles.header}>Ustawienia konta</h1>
          <p>Tu możesz zmienić ustawienia swojego konta.</p>
        </div>

        {/* <article class={styles.setting}>
          <h2 class={styles.setting_header}>Status aktywności</h2>
          <p class={styles.setting_description}>Obecny status: {AppState.store.userInfo?.is_active ? 'Aktywny' : 'Nieaktywny'}</p>
          <button class={styles.setting_button}>Zmień status</button>
        </article> */}
        <form class={styles.setting} method='post' onSubmit={async (e) => {
          e.preventDefault();
          const formData = new FormData(e.currentTarget);
          const chosenColumn = formData.get('chosen_column');
          if (chosenColumn) {
            const success = await changeChosenColumn(Number(chosenColumn));
            if (success) {
              alert('Pomyślnie zmieniono kolumnę.');

              AppState.setStore('userInfo', "chosen_column", Number(chosenColumn));
            } else {
              alert('Wystąpił błąd podczas zmiany kolumny. Spróbuj ponownie.');
            }
          }
        }}>

          <h2 class={styles.setting_header}>Wybrana kolumna</h2>
          <p>Zmianę kolumny wykonuj wyłącznie w uzasadnionych przypadkach, aby uniknąć ograniczeń konta.</p>
          <p class={styles.setting_description}>Obecna kolumna: {AppState.store.userInfo?.chosen_column === null ? 'Nie wybrano kolumny' : `Kolumna ${AppState.store.userInfo?.chosen_column}`}</p>
          <select class={styles.column_select} name="chosen_column" id="chosen_column">
            <For each={columnOptions}>{(option) => (
              <option value={option.value} selected={AppState.store.userInfo?.chosen_column === option.value}>{option.label}</option>
            )}</For>
          </select>
          <button class={styles.setting_button}>Zmień kolumnę</button>
        </form>



      </section>
    </main>
  );
}
