import { PageTitle } from "#components/PageTitle";
import * as stylex from "@stylexjs/stylex";
import { createMemo, createResource, createSignal, Show, useContext } from "solid-js";
import { A } from "@solidjs/router";
import { AppContext } from "../appContext";
import { Button } from "../components/Button";
import { UserCard } from "../components/UserCard";
import { TimetablePreview } from "../components/Timetable/TimetablePreview";
import { TimetablePicker } from "../components/Timetable/TimetablePicker";
import { readGuestSelection, resolveGuestSelection, saveGuestSelection } from "../components/Timetable/guestSelection";
import { pageStyles } from "../styles/pageStyles";
import { GOOGLE_CALENDAR_URL, PK_WEBSITE_URL } from "../constants";
import { getAuthURL, getClasses, getColumns, getOverrides, getPublicTimetable } from "../services/pkplaner";
import { attrs, setAccessToken } from "../utils";

const MOBILE_MAX = "@media (max-width: 640px)" as const;

const styles = stylex.create({
	topRow: {
		display: "grid",
		gridTemplateColumns: {
			default: "minmax(0, 1fr) auto",
			"@media (max-width: 1100px)": "minmax(0, 1fr)",
		},
		alignItems: "start",
		gap: "16px 24px",
		paddingInline: "16px",
	},
	description: {
		display: "flex",
		flexDirection: "column",
		minWidth: 0,
		gap: "8px",
		color: "#576475",
		fontSize: "0.9rem",
	},
	header: {
		fontSize: "1.65rem",
		fontWeight: 600,
		letterSpacing: "-0.025em",
		color: "#0b2a45",
		lineHeight: 1.25,
	},
	interactionsContainer: {
		gap: "1em",
		alignItems: {
			[MOBILE_MAX]: "flex-start",
			default: "center",
		},
		display: "flex",
		flexDirection: {
			[MOBILE_MAX]: "column",
			default: "row",
		},
		justifyContent: "space-between",
		minWidth: 0,
	},
	authButtons: {
		gap: "0.75em",
		display: { default: "flex", [MOBILE_MAX]: "grid" },
		gridTemplateColumns: { [MOBILE_MAX]: "repeat(2, minmax(0, 1fr))" },
		width: { default: "auto", [MOBILE_MAX]: "100%" },
		justifyContent: "flex-start",
		flexWrap: "wrap",
		flexShrink: 0,
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
	const [columns] = createResource(getColumns);
	const [browseSelection, setBrowseSelection] = createSignal(readGuestSelection());
	const browseColumn = createMemo(() =>
		resolveGuestSelection(
			columns.error ? [] : (columns() ?? []),
			browseSelection() ?? (selectedGroup() ? Number(selectedGroup()) : undefined)
		)
	);
	const viewingOwnPlan = () => !!userInfo() && selectedGroup() !== false && String(browseColumn()) === selectedGroup();
	const [overrides] = createResource(() => viewingOwnPlan() && selectedGroup(), getOverrides);
	const previewSource = () =>
		browseColumn() === undefined ? false : viewingOwnPlan() ? `user:${selectedGroup()}` : `public:${browseColumn()}`;
	const [classes] = createResource(previewSource, (source) =>
		source.startsWith("user:") ? getClasses() : getPublicTimetable(Number(source.slice(7)))
	);
	const previewGroup = () =>
		columns.error ? undefined : columns()?.find((column) => column.column === browseColumn());
	function selectBrowseColumn(column: number) {
		setBrowseSelection(column);
		saveGuestSelection(column);
	}

	return (
		<main {...attrs(stylex.props(pageStyles.page))}>
			<PageTitle />

			<section {...attrs(stylex.props(pageStyles.section))}>
				<div {...attrs(stylex.props(styles.topRow))}>
					<div {...attrs(stylex.props(styles.description))}>
						<h1 {...attrs(stylex.props(styles.header))}>
							{userInfo() ? "Plany zajęć" : "Plan z uczelni. W Twoim kalendarzu."}
						</h1>

						<Show when={!userInfo()} fallback={<p>Twój tygodniowy rozkład i synchronizacja z Kalendarzem Google.</p>}>
							<p>
								Aplikacja pozwalająca na śledzenie rozkładów zajęć{" "}
								<a href={PK_WEBSITE_URL}>Politechniki Krakowskiej im. Tadeusza Kościuszki</a>. Po zalogowaniu się za
								pomocą konta Google, podaniu swojego wydziału, kierunku, a także grupy, aplikacja automatycznie będzie
								wprowadzać zmiany w <a href={GOOGLE_CALENDAR_URL}>kalendarz Google</a> prosto z rozkładów zajęć .xls
								zamieszczonych na stronie PK.
							</p>
						</Show>
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
									<UserCard
										compact
										userInfo={info()}
										groupLabel={
											columns.error
												? undefined
												: columns()?.find((column) => String(column.column) === selectedGroup())?.label
										}
									/>
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
				</div>
				<TimetablePreview
					expand
					headerAction={
						<Show when={!columns.error && (columns()?.length ?? 0) > 0}>
							<TimetablePicker
								compact
								columns={columns() ?? []}
								selected={browseColumn()}
								onChange={selectBrowseColumn}
							/>
						</Show>
					}
					groupLabel={previewGroup() ? `${previewGroup()!.category} · ${previewGroup()!.label}` : undefined}
					classes={classes.error ? undefined : classes()}
					overrides={viewingOwnPlan() && !overrides.error ? overrides() : undefined}
					loading={classes.loading || columns.loading || (viewingOwnPlan() && overrides.loading)}
					error={classes.error || columns.error || (viewingOwnPlan() && overrides.error)}
				/>
				<p {...attrs(stylex.props(pageStyles.footer))}>
					<A href="/privacy-policy">Polityka prywatności</A>
				</p>
			</section>
		</main>
	);
}
