import { UserInfoResponse } from "../../constants";
import styles from "./UserCard.module.css";

type UserCardProps = {
	userInfo: UserInfoResponse;
};

export function UserCard(props: UserCardProps) {
	return (
		<article class={styles.card}>
			<img class={styles.avatar} src={props.userInfo.picture} alt={`${props.userInfo.name}'s profile picture`} />

			<div class={styles.details}>
				<p class={styles.name}>{props.userInfo.name}</p>

				<p class={styles.email}>{props.userInfo.email}</p>

				<div class={styles.metaRow}>
					<span
						class={`${styles.metaLabel} ${styles.status} ${
							props.userInfo.is_active ? styles.statusActive : styles.statusInactive
						}`}
					>
						{props.userInfo.is_active ? "Aktywny" : "Nieaktywny"}
					</span>
					<span class={styles.metaLabel}>
						{props.userInfo.chosen_column === null ? "Nie wybrano kolumny" : `Kolumna ${props.userInfo.chosen_column}`}
					</span>
				</div>
			</div>
		</article>
	);
}
