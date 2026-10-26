import * as stylex from "@stylexjs/stylex";
import { attrs } from "../../utils";
import { weekdays } from "./timetable";
import { timetableStyles as styles } from "./timetableStyles";

export function TimetableDayNavigation(props: { selected: number; onChange: (day: number) => void }) {
	return (
		<nav {...attrs(stylex.props(styles.navigation))} aria-label="Wybór dnia planu">
			<button
				{...attrs(stylex.props(styles.arrow))}
				type="button"
				aria-label="Poprzedni dzień"
				onClick={() => props.onChange((props.selected + 4) % 5)}
			>
				←
			</button>
			<strong aria-live="polite">{weekdays[props.selected].label}</strong>
			<button
				{...attrs(stylex.props(styles.arrow))}
				type="button"
				aria-label="Następny dzień"
				onClick={() => props.onChange((props.selected + 1) % 5)}
			>
				→
			</button>
		</nav>
	);
}
