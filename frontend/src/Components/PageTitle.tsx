import { Title } from "@solidjs/meta";

export function PageTitle(props: { title?: string }) {
	const fullTitle = () => (props.title ? `${props.title} | PlanerPK ` : "PlanerPK ");
	return <Title>{fullTitle()}</Title>;
}
