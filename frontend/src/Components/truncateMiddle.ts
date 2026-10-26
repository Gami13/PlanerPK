// Keep both ends of a title, using the actual rendered text width.
export function middleEllipsis(text: string, width: number, measure: (text: string) => number): string {
	if (measure(text) <= width) return text;
	if (measure("…") > width) return "";
	const characters = Array.from(text);
	const candidate = (count: number) => {
		const head = Math.ceil(count / 2);
		const tail = Math.floor(count / 2);
		return `${characters.slice(0, head).join("")}…${tail ? characters.slice(-tail).join("") : ""}`;
	};
	let low = 0;
	let high = characters.length - 1;
	while (low < high) {
		const middle = Math.ceil((low + high) / 2);
		if (measure(candidate(middle)) <= width) low = middle;
		else high = middle - 1;
	}
	return candidate(low);
}
