import type { FeatureNode, ProjectInfo } from "@/features/product-setup/types/productSetup.types";

const normalizedText = (value?: string) => value?.trim() ?? "";

/**
 * Produces a stable representation of the project fields that the setup flow
 * persists. Server-generated fields, such as codes and display order, are
 * deliberately excluded so they do not make a clean form appear dirty.
 */
export function projectSnapshot(projects: ProjectInfo[]): string {
  return JSON.stringify(
    projects.map((project) => ({
      id: String(project.project_id ?? project.id ?? ""),
      name: normalizedText(project.project_name),
      description: normalizedText(project.description),
      isActive: project.is_active ?? true,
    })),
  );
}

/**
 * Produces a stable representation of the editable hierarchy fields. IDs and
 * parent IDs are retained because they identify additions, removals, and moves.
 */
export function featureSnapshot(features: FeatureNode[]): string {
  return JSON.stringify(
    features.map((feature) => ({
      id: String(feature.category_id ?? feature.id),
      projectId: String(feature.project_id),
      parentId:
        feature.parent_id === null || feature.parent_id === undefined
          ? null
          : String(feature.parent_id),
      name: normalizedText(feature.category_name),
      description: normalizedText(feature.description),
      isActive: feature.is_active ?? true,
    })),
  );
}
