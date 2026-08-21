import type { FeatureNode, ProjectInfo } from "../types/productSetup.types";

export function toCode(input: string): string {
  return input
    .trim()
    .toUpperCase()
    .replace(/[^\p{L}\p{M}\p{N}_]/gu, "_")
    .replace(/_+/g, "_")
    .replace(/^_+|_+$/g, "");
}

export function validateProjectCodeUniqueness(projects: ProjectInfo[]): { valid: boolean; duplicates: string[] } {
  const seen = new Set<string>();
  const duplicates = new Set<string>();

  for (const p of projects) {
    const code = toCode(p.project_code ?? "");
    if (!code) continue;
    if (seen.has(code)) {
      duplicates.add(code);
    }
    seen.add(code);
  }

  return {
    valid: duplicates.size === 0,
    duplicates: Array.from(duplicates),
  };
}

export function validateFeatureCodeUniqueness(features: FeatureNode[]): { valid: boolean; duplicates: string[] } {
  const seenPerProject = new Map<string | number, Set<string>>();
  const duplicates = new Set<string>();

  for (const f of features) {
    const code = toCode(f.category_code ?? "");
    if (!code) continue;
    
    let projSet = seenPerProject.get(f.project_id);
    if (!projSet) {
      projSet = new Set<string>();
      seenPerProject.set(f.project_id, projSet);
    }

    if (projSet.has(code)) {
      duplicates.add(code);
    }
    projSet.add(code);
  }

  return {
    valid: duplicates.size === 0,
    duplicates: Array.from(duplicates),
  };
}

export function checkCircularParent(features: FeatureNode[], featureId: string | number, targetParentId: string | number | null): boolean {
  if (!targetParentId) return false;
  if (featureId === targetParentId) return true;

  const featMap = new Map<string | number, FeatureNode>();
  for (const f of features) {
    const id = f.category_id ?? f.id;
    featMap.set(id, f);
  }

  let current: string | number | null = targetParentId;
  const visited = new Set<string | number>();

  while (current !== null) {
    if (current === featureId) return true; // Circular detected!
    if (visited.has(current)) break;
    visited.add(current);

    const node = featMap.get(current);
    if (!node) break;
    current = node.parent_id;
  }

  return false;
}
