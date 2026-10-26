import { PageTitle } from "#components/PageTitle";
import * as stylex from "@stylexjs/stylex";
import { createResource, Show, useContext } from "solid-js";
import { A } from "@solidjs/router";
import { AppContext } from "../appContext";
import { Button } from "../Components/Button";
import { UserCard } from "../Components/UserCard";
import { TimetablePreview } from "../Components/Timetable/TimetablePreview";
import { GOOGLE_CALENDAR_URL, PK_WEBSITE_URL } from "../constants";
import { getAuthURL, getClasses, getColumns, getOverrides } from "../services/pkplaner";
import { attrs, setAccessToken } from "../utils";

const MOBILE_MAX = "@media (max-width: 33.75em)" as const;
const SMALL_SCREEN = "@media (min-width: 38em)" as const;

const styles = stylex.create({
	main: {
		padding: "2em",
		gap: "2em",
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
		width: "100%",
	},
	basicSection: {
		width: "100%",
		gap: "1em",
		alignItems: "center",
		display: "flex",
		flexDirection: "column",
	},
	description: {
		display: "flex",
		flexDirection: "column",
		maxWidth: "48em",
	},
	header: {
		fontSize: "2.8em",
		fontWeight: 500,
	},
	interactionsContainer: {
		gap: "1em",
		alignItems: {
			[MOBILE_MAX]: "stretch",
			default: "center",
		},
		display: "flex",
		flexDirection: {
			[MOBILE_MAX]: "column",
			default: "row",
		},
		justifyContent: "space-between",
		maxWidth: "48em",
		width: "100%",
	},
	authButtons: {
		gap: "0.75em",
		display: "flex",
		justifyContent: {
			[MOBILE_MAX]: "center",
			default: null,
		},
		width: {
			[MOBILE_MAX]: "100%",
			default: null,
		},
	},

	imageContainer: {
		transition: "width 0.3s ease, border-width 0.3s ease",
		display: "inline-block",
		transformOrigin: {
			[SMALL_SCREEN]: "center center",
			default: null,
		},
		maxWidth: "95vw",
		width: {
			default: "min(100%, 48em)",
			":hover": {
				[SMALL_SCREEN]: "100%",
				default: null,
			},
		},
	},
	googleCalendarPreview: {
		borderColor: "#e4e8ee",
		borderRadius: "8px",
		borderStyle: "solid",
		borderWidth: "1px",
		boxSizing: "border-box",
		height: "auto",
		maxWidth: "100%",
	},
});

export function logOut() {
	setAccessToken("");
	window.location.reload();
}

export function Home() {
	const AppState = useContext(AppContext);
	const [authURL] = createResource(getAuthURL);
	const userInfo = () => AppState.store.userInfo;
	const selectedGroup = () => (userInfo()?.chosenColumn ? String(userInfo()!.chosenColumn) : false);
	const [classes] = createResource(selectedGroup, getClasses);
	const [overrides] = createResource(selectedGroup, getOverrides);
	const [columns] = createResource(selectedGroup, getColumns);

	return (
		<main {...attrs(stylex.props(styles.main))}>
			<PageTitle />

			<section {...attrs(stylex.props(styles.basicSection))}>
				<div {...attrs(stylex.props(styles.description))}>
					<h1 {...attrs(stylex.props(styles.header))}>Google Calendar Timetable</h1>

					<p>
						Aplikacja pozwalająca na śledzenie rozkładów zajęć{" "}
						<a href={PK_WEBSITE_URL}>Politechniki Krakowskiej im. Tadeusza Kościuszki</a>. Po zalogowaniu się za pomocą
						konta Google, podaniu swojego wydziału, kierunku, a także grupy, aplikacja automatycznie będzie wprowadzać
						zmiany w <a href={GOOGLE_CALENDAR_URL}>kalendarz Google</a> prosto z rozkładów zajęć .xls zamieszczonych na
						stronie PK.
					</p>
				</div>
				<div {...attrs(stylex.props(styles.interactionsContainer))}>
					<Show
						when={userInfo()}
						fallback={
							<Button isPrimary={true} isVanillaLink={true} href={authURL()}>
								Zarejestruj się z Google
							</Button>
						}
					>
						{(info) => (
							<>
								<UserCard userInfo={info()} />
								<div {...attrs(stylex.props(styles.authButtons))}>
									<Button isPrimary={true} href="/settings">
										Ustawienia
									</Button>

									<Button isPrimary={false} onClick={logOut}>
										Wyloguj
									</Button>
								</div>
							</>
						)}
					</Show>
				</div>

				<Show
					when={userInfo()}
					fallback={
						<div {...attrs(stylex.props(styles.imageContainer))}>
							<img
								{...attrs(stylex.props(styles.googleCalendarPreview))}
								src="/google-calendar-preview.png"
								alt="Przykładowy plan zajęć w Kalendarzu Google"
							/>
						</div>
					}
				>
					<TimetablePreview
						groupLabel={
							columns.error ? undefined : columns()?.find((column) => String(column.column) === selectedGroup())?.label
						}
						classes={classes.error ? undefined : classes()}
						overrides={overrides.error ? undefined : overrides()}
						loading={classes.loading || overrides.loading}
						error={classes.error || overrides.error}
						noGroup={!selectedGroup()}
					/>
				</Show>
				<p>
					<A href="/privacy-policy">Polityka prywatności</A>
				</p>
			</section>
		</main>
	);
}
