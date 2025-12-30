import { Title } from "@solidjs/meta";
import { Show } from "solid-js";

export function PageTitle(props: { title?: string }) {
	return (
		<Show when={props.title != null} fallback={<Title>PlanerPK </Title>}>
			<Title>{props.title} | PlanerPK </Title>
		</Show>
	);
}
