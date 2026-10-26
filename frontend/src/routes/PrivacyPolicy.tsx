import { A } from "@solidjs/router";
import * as stylex from "@stylexjs/stylex";
import { PageTitle } from "../Components/PageTitle";
import { AppHeader } from "../Components/AppHeader";
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
			<AppHeader />
			<h1 {...attrs(stylex.props(styles.heading))}>Polityka prywatności</h1>
			<p {...attrs(stylex.props(styles.muted))}>Ostatnia aktualizacja: 23 września 2026 r.</p>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>1. Administrator</h2>
				<p>
					PlanerPK („Serwis") jest prowadzony przez twórców Serwisu. Serwis synchronizuje rozkład zajęć Politechniki
					Krakowskiej z Kalendarzem Google. Kontakt: <a href="mailto:gami@gami.ovh">gami@gami.ovh</a>.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>2. Przetwarzane dane i cel</h2>
				<p>Serwis przetwarza wyłącznie dane niezbędne do działania:</p>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>profil Google (e-mail, imię, zdjęcie) — wyświetlanie konta;</li>
					<li>dane uwierzytelniające OAuth — logowanie i dostęp do API Google w imieniu użytkownika;</li>
					<li>
						wydarzenia Kalendarza Google — synchronizacja rozkładu (Serwis nie modyfikuje wydarzeń, których sam nie
						utworzył);
					</li>
					<li>wybraną grupę zajęć i ustawienia konta.</li>
				</ul>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>3. Wykorzystanie danych Google</h2>
				<p>
					Dane Google nie są sprzedawane, wykorzystywane do reklam, profilowania ani do trenowania modeli AI.
					Korzystanie odbywa się zgodnie z{" "}
					<a href="https://developers.google.com/terms/api-services-user-data-policy">
						Google API Services User Data Policy
					</a>
					, w tym z wymogami Limited Use.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>4. Przechowywanie i bezpieczeństwo</h2>
				<p>
					Dane są przechowywane w bazie danych na serwerze VPS administrowanym przez twórców Serwisu (Europa).
					Komunikacja z Serwisem odbywa się przez HTTPS. Dostęp do serwera jest ograniczony do administratorów.
				</p>
				<p>
					Dane przechowywane są tak długo, jak konto jest aktywne. Po usunięciu konta wszystkie powiązane dane (profil,
					dane uwierzytelniające, ustawienia, nadpisania zajęć) są trwale usuwane z bazy danych.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>5. Podmioty trzecie</h2>
				<p>
					Serwis komunikuje się z Google OAuth, Google Calendar i opcjonalnie Google Places (wyłącznie nazwy budynków).
					Ruch sieciowy przechodzi przez Cloudflare (CDN/proxy). Serwis nie korzysta z narzędzi analitycznych ani
					reklamowych. Przeglądarka może łączyć się z Google Fonts.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>6. Prawa użytkownika i usunięcie danych</h2>
				<p>
					Użytkownik ma prawo dostępu, sprostowania i usunięcia swoich danych. Zgodę na dostęp do Kalendarza Google
					można cofnąć w <a href="https://myaccount.google.com/permissions">ustawieniach konta Google</a>.
				</p>
				<p>
					Aby usunąć konto i wszystkie powiązane dane z Serwisu, napisz na{" "}
					<a href="mailto:gami@gami.ovh">gami@gami.ovh</a>. Usunięcie danych Serwisu nie usuwa wydarzeń już zapisanych w
					Kalendarzu Google.
				</p>
			</section>

			<p>
				<A {...attrs(stylex.props(styles.link))} href="/">
					Wróć do strony głównej
				</A>
			</p>
		</main>
	);
}
