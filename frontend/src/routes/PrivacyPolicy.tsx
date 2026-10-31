import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { PageTitle } from "../Components/PageTitle";
import { attrs } from "../utils";

const styles = stylex.create({
	main: {
		color: "#0f172a",
		display: "flex",
		flexDirection: "column",
		gap: "1.25em",
		margin: "0 auto",
		maxWidth: "56em",
		padding: "2em",
		width: "100%",
	},
	heading: {
		fontSize: "1.65rem",
		fontWeight: 600,
		letterSpacing: "-0.025em",
	},
	section: {
		display: "flex",
		flexDirection: "column",
		gap: "0.7em",
	},
	sectionHeading: {
		fontSize: "1.05rem",
		fontWeight: 600,
	},
	list: {
		display: "flex",
		flexDirection: "column",
		gap: "0.45em",
		paddingInlineStart: "1.5em",
	},
	link: {
		color: "#245a85",
	},
	muted: {
		color: "#546171",
		fontSize: "0.9em",
	},
});

export function PrivacyPolicy() {
	return (
		<main {...attrs(stylex.props(styles.main))}>
			<PageTitle title="Polityka prywatności" />
			<h1 {...attrs(stylex.props(styles.heading))}>Polityka prywatności</h1>
			<p {...attrs(stylex.props(styles.muted))}>Ostatnia aktualizacja: 31 października 2026 r.</p>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>1. Administrator i kontakt</h2>
				<p>
					Administratorem serwisu PlanerPK („Serwis"), służącego do synchronizacji planu zajęć Politechniki
					Krakowskiej z Kalendarzem Google, są jego twórcy. Kontakt we wszystkich sprawach związanych z danymi:{" "}
					<a href="mailto:gami@gami.ovh">gami@gami.ovh</a>.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>2. Przetwarzane dane i wykorzystanie API Google</h2>
				<p>Serwis przetwarza wyłącznie dane niezbędne do działania:</p>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>
						<strong>Profil Google (e-mail, imię, zdjęcie) oraz tokeny OAuth</strong> — do logowania, utrzymania sesji
						i autoryzacji zapytań do API Google.
					</li>
					<li>
						<strong>
							Zakres <code>calendar.app.created</code>
						</strong>{" "}
						— wyłącznie do utworzenia dedykowanego kalendarza „PlanerPK" i zarządzania w nim wydarzeniami. Serwis
						nie ma dostępu do innych kalendarzy użytkownika.
					</li>
					<li>
						<strong>Konfiguracja planu</strong> — wybrana grupa, ustawienia konta oraz własne modyfikacje zajęć (np.
						ukryte przedmioty, dodane grupy, zmienione godziny).
					</li>
				</ul>
				<p>
					Dane z API Google nie są sprzedawane ani wykorzystywane do reklam, profilowania czy trenowania modeli AI.
					Korzystanie z danych odbywa się zgodnie z zasadami{" "}
					<a href="https://developers.google.com/terms/api-services-user-data-policy">
						Google API Services User Data Policy
					</a>
					, w tym z wymogami <em>Limited Use</em>.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>3. Przechowywanie, bezpieczeństwo i podmioty trzecie</h2>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>
						Dane są przechowywane na zabezpieczonym serwerze VPS w Europie, a komunikacja odbywa się wyłącznie przez
						HTTPS.
					</li>
					<li>
						Serwis korzysta z usług Google (OAuth, Calendar, Fonts oraz opcjonalnie Places do nazw budynków) oraz
						Cloudflare (ochrona sieci i CDN). Nie używamy narzędzi analitycznych ani reklamowych.
					</li>
				</ul>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>4. Prawa użytkownika i usuwanie danych</h2>
				<p>
					Dane przechowujemy przez czas aktywności konta. Użytkownik ma prawo dostępu do swoich danych, ich
					sprostowania, ograniczenia przetwarzania oraz trwałego usunięcia.
				</p>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>
						<strong>Cofnięcie dostępu:</strong> Autoryzację można w każdej chwili wycofać w{" "}
						<a href="https://myaccount.google.com/permissions">ustawieniach konta Google</a>.
					</li>
					<li>
						<strong>Usunięcie konta:</strong> Wyślij wiadomość na{" "}
						<a href="mailto:gami@gami.ovh">gami@gami.ovh</a> — trwale usuniemy Twój profil, tokeny i ustawienia z
						bazy danych. Wydarzenia zapisane wcześniej w Kalendarzu Google pozostają w nim do momentu ręcznego
						usunięcia kalendarza „PlanerPK" przez użytkownika.
					</li>
				</ul>
			</section>

			<p>
				<A {...attrs(stylex.props(styles.link))} href="/">
					Wróć do strony głównej
				</A>
			</p>
		</main>
	);
}
