import type { Project } from "@/lib/api";

export const NO_PROJECT_OPTION = {
  value: "",
  label: "No project",
} as const;

export const NO_PROJECT_FIELD_HINT =
  "Optional. You can assign a project later from the resource settings.";

export function projectSelectOptions(projects: Project[]) {
  const userProjects = projects.filter((p) => !p.system);
  return [
    NO_PROJECT_OPTION,
    ...userProjects.map((p) => ({ value: p.id, label: p.name })),
  ];
}
