import { ColumnGroup } from "./constants";

export function setAccessToken(token: string) {
	document.cookie = `access_token=${token}; path=/; max-age=604800; samesite=strict`;
}
export function getAccessToken(): string | null {
	const match = document.cookie.match(new RegExp("(^| )access_token=([^;]+)"));
	return match ? match[2] : null;
}

//StyleX utility to convert props to Solid-compatible attributes
//example use <div {...attrs(stylex.props(...styles))} />

interface StyleXProps {
	className?: string | null;
	"data-style-src"?: string | null;
	style?: Record<string, string | number> | null;
}

interface SolidAttrs {
	class?: string;
	style?: string;
	"data-style-src"?: string;
}

export function attrs({ className, "data-style-src": dataStyleSrc, style }: StyleXProps): SolidAttrs {
	const result: SolidAttrs = {};
	// Convert className to class
	if (className != null && className !== "") {
		result.class = className;
	}
	// Convert style object to string
	if (style != null && Object.keys(style).length > 0) {
		result.style = Object.keys(style)
			.map((key) => `${key}:${style[key]};`)
			.join("");
	}
	if (dataStyleSrc != null && dataStyleSrc !== "") {
		result["data-style-src"] = dataStyleSrc;
	}
	return result;
}

export function getGroupByColumnIndex(
	columns: ColumnGroup[] | undefined,
	columnIndex: number
): ColumnGroup | undefined {
	return columns?.find((col) => col.column === columnIndex);
}
