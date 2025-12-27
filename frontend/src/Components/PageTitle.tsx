import { Title } from "@solidjs/meta";

export function PageTitle(props: { title?: string }) {
	if (!props.title) {
		return <Title>PlanerPK </Title>
	}
	return <Title>{props.title} | PlanerPK </Title>
}