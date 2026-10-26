import type { ClassInfo } from "../../gen/pkplaner/v1/pkplaner_pb";

export const defaultTitleTemplate = "{ClassTitle} ({ClassType}) - {ClassTeacher}";

export const titlePlaceholders = [
	"ClassTitle",
	"ClassType",
	"ClassTeacher",
	"ClassAudience",
	"ClassRecurrence",
] as const;

type TemplateClass = Pick<ClassInfo, "name" | "type" | "teacher" | "audience" | "recurrence">;

export function formatClassTitle(template: string, classInfo: TemplateClass): string {
	const values: Record<string, string> = {
		ClassTitle: classInfo.name ?? "",
		ClassType: classInfo.type ?? "",
		ClassTeacher: classInfo.teacher ?? "",
		ClassAudience: classInfo.audience ?? "",
		ClassRecurrence: classInfo.recurrence ?? "",
	};
	// One pass: replacement text is literal and is never interpreted as another placeholder.
	return template.replace(/\{([^{}]+)\}/g, (placeholder, key: string) =>
		Object.hasOwn(values, key) ? values[key] : placeholder
	);
}
