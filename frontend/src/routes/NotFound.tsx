import { PageTitle } from "../components/PageTitle";
import * as stylex from "@stylexjs/stylex";
import { Button } from "../components/Button";
import { pageStyles } from "../styles/pageStyles";
import { attrs } from "../utils";

export function NotFound() {
	return (
		<main {...attrs(stylex.props(pageStyles.page))}>
			<PageTitle title="Nic tu nie ma :(" />
			<section {...attrs(stylex.props(pageStyles.intro))}>
				<h1 {...attrs(stylex.props(pageStyles.heading))}>Nie znaleziono strony</h1>
				<p {...attrs(stylex.props(pageStyles.muted))}>Ten adres nie prowadzi do żadnej strony w PlanerPK.</p>
				<div>
					<Button href="/" isPrimary>
						Wróć do planu
					</Button>
				</div>
			</section>
		</main>
	);
}
