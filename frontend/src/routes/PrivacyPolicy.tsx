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
		fontSize: "2.4em",
		fontWeight: 500,
	},
	section: {
		display: "flex",
		flexDirection: "column",
		gap: "0.7em",
	},
	sectionHeading: {
		fontSize: "1.35em",
		fontWeight: 600,
	},
	list: {
		display: "flex",
		flexDirection: "column",
		gap: "0.45em",
		paddingInlineStart: "1.5em",
	},
	link: {
		color: "#039be5",
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
			<p {...attrs(stylex.props(styles.muted))}>Ostatnia aktualizacja: 24 sierpnia 2026 r.</p>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>1. Informacje ogólne</h2>
				<p>
					PlanerPK (dalej: „Serwis”) pomaga użytkownikom przeglądać rozkłady zajęć Politechniki Krakowskiej i synchronizować
					wybraną grupę zajęć z własnym Kalendarzem Google.
				</p>
				<p>
					Serwis jest prowadzony przez twórców projektu PlanerPK. W sprawach dotyczących prywatności można skontaktować się
					pod adresem <a href="mailto:gamiwastaken@gmail.com">gamiwastaken@gmail.com</a>.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>2. Dane, które przetwarzamy</h2>
				<p>W zależności od używanych funkcji Serwis może przetwarzać:</p>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>
						podstawowe informacje z konta Google: adres e-mail, imię i nazwisko, imię oraz adres URL zdjęcia profilowego;
					</li>
					<li>
						identyfikatory i tokeny OAuth Google, w tym token odświeżania, potrzebne do uwierzytelnienia użytkownika i
						synchronizacji kalendarza;
					</li>
					<li>
						wybraną kolumnę lub grupę zajęć oraz ustawienia użytkownika dotyczące aktywności konta i zmian w zajęciach;
					</li>
					<li>
						informacje o wydarzeniach Kalendarza Google potrzebne do rozpoznania wydarzeń utworzonych przez Serwis i
						porównania ich z aktualnym rozkładem;
					</li>
					<li>
						informacje o nazwach i lokalizacjach budynków używane do opcjonalnego wyszukiwania miejsc na mapie.
					</li>
				</ul>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>3. Jak wykorzystujemy dane Google</h2>
				<p>Dane Google wykorzystujemy wyłącznie do działania funkcji wybranych przez użytkownika:</p>
				<ul {...attrs(stylex.props(styles.list))}>
					<li>informacje profilowe służą do utworzenia i wyświetlenia konta użytkownika w Serwisie;</li>
					<li>
						dostęp do Kalendarza Google służy do odczytania wydarzeń z głównego kalendarza, znalezienia wydarzeń
						utworzonych przez Serwis oraz utworzenia, aktualizacji lub usunięcia tych wydarzeń podczas synchronizacji;
					</li>
					<li>Serwis nie modyfikuje wydarzeń, których sam wcześniej nie utworzył;</li>
					<li>
						dane Google nie są sprzedawane, wykorzystywane do reklam, profilowania ani przekazywane w celu niezwiązanym z
						funkcją synchronizacji;
					</li>
					<li>
						Serwis nie wykorzystuje danych Google do trenowania modeli sztucznej inteligencji.
					</li>
				</ul>
				<p>
					Korzystanie z danych Google odbywa się zgodnie z wymaganiami Google API Services User Data Policy, w tym z
					wymogami Limited Use.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>4. Kalendarz Google</h2>
				<p>
					Po udzieleniu zgody Serwis korzysta z uprawnień Google Calendar do synchronizacji rozkładu. Serwis może odczytać
					wydarzenia z głównego kalendarza, aby rozpoznać własne wydarzenia i uniknąć ich duplikowania. Nowe wydarzenia
					tworzone przez Serwis zawierają oznaczenie pozwalające odróżnić je od pozostałych wydarzeń użytkownika.
				</p>
				<p>
					Użytkownik może cofnąć zgodę w ustawieniach swojego konta Google. Cofnięcie zgody może uniemożliwić dalszą
					synchronizację, ale nie usuwa automatycznie wydarzeń już utworzonych w Kalendarzu Google.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>5. Przechowywanie i udostępnianie danych</h2>
				<p>
					Dane konta, tokeny OAuth oraz ustawienia użytkownika są przechowywane na serwerze Serwisu w bazie danych
					SQLite. Tokeny są używane wyłącznie do uwierzytelnienia żądań do Google w imieniu użytkownika i synchronizacji
					jego kalendarza.
				</p>
				<p>
					Serwis korzysta z usług Google API, w tym Google OAuth, Google Calendar i opcjonalnie Google Places. Nazwy
					budynków mogą być przekazywane do Google Places wyłącznie w celu uzupełnienia informacji o lokalizacji; dane konta
					Google nie są w tym celu przekazywane.
				</p>
				<p>
					Serwis nie korzysta z narzędzi analitycznych ani reklamowych. Aplikacja korzysta z Google Fonts, dlatego przeglądarka
					może nawiązać bezpośrednie połączenie z usługą Google Fonts w celu pobrania kroju pisma.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>6. Okres przechowywania i usunięcie danych</h2>
				<p>
					Dane są przechowywane tak długo, jak długo konto jest używane lub jest potrzebne do obsługi synchronizacji. Aby
					poprosić o usunięcie konta i danych przechowywanych przez Serwis, napisz na adres
					{" "}
					<a href="mailto:gamiwastaken@gmail.com">gamiwastaken@gmail.com</a>. Prośba powinna zawierać adres e-mail
					konta Google używanego w Serwisie.
				</p>
				<p>
					Usunięcie danych Serwisu nie usuwa wydarzeń zapisanych w Kalendarzu Google. Wydarzenia można usunąć samodzielnie
					w Kalendarzu Google lub poprosić o pomoc w wiadomości dotyczącej usunięcia konta.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>7. Bezpieczeństwo</h2>
				<p>
					Stosujemy rozsądne środki techniczne i organizacyjne mające chronić dane przed nieuprawnionym dostępem,
					utratą lub zmianą. Żadna usługa internetowa nie gwarantuje jednak całkowitego bezpieczeństwa danych.
				</p>
			</section>

			<section {...attrs(stylex.props(styles.section))}>
				<h2 {...attrs(stylex.props(styles.sectionHeading))}>8. Zmiany polityki</h2>
				<p>
					Polityka może być aktualizowana, gdy zmieni się sposób działania Serwisu lub wymagania prawne. Aktualna wersja
					jest zawsze dostępna na tej stronie.
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
