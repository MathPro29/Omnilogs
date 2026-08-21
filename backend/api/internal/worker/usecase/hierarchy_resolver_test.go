package usecase

import (
	"testing"

	"omnilogs-api/models"
)

func TestBestAutomaticFeatureMatch(t *testing.T) {
	features := []models.ProjectFeature{
		{CategoryID: 10, CategoryCode: "water_bills", CategoryName: "Water bills"},
		{CategoryID: 20, CategoryCode: "other_bills", CategoryName: "Other bills"},
		{CategoryID: 30, CategoryCode: "members", CategoryName: "Residents"},
	}

	tests := []struct {
		name  string
		hints []string
		want  int
	}{
		{name: "exact code", hints: []string{"water_bills"}, want: 10},
		{name: "near spelling", hints: []string{"water_bill"}, want: 10},
		{name: "nested route tokens", hints: []string{"/api/bill_others/bot_settings/add"}, want: 20},
		{name: "no match", hints: []string{"/api/health/check"}, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matched := bestAutomaticFeatureMatch(features, tt.hints)
			if tt.want == 0 {
				if matched != nil {
					t.Fatalf("unexpected match: %s", matched.CategoryCode)
				}
				return
			}
			if matched == nil || matched.CategoryID != tt.want {
				t.Fatalf("match = %#v, want category %d", matched, tt.want)
			}
		})
	}
}

func TestBestAutomaticFeatureMatchRejectsAmbiguousResult(t *testing.T) {
	features := []models.ProjectFeature{
		{CategoryID: 10, CategoryCode: "billing", CategoryName: "Billing"},
		{CategoryID: 20, CategoryCode: "bills", CategoryName: "Bills"},
	}
	if matched := bestAutomaticFeatureMatch(features, []string{"/api/billing/bills/add"}); matched != nil {
		t.Fatalf("ambiguous hierarchy must use product/environment fallback: %#v", matched)
	}
}

func TestBestAutomaticFeatureMatchRejectsSameFeatureCodeAcrossProjects(t *testing.T) {
	features := []models.ProjectFeature{
		{CategoryID: 10, ProjectID: 100, CategoryCode: "members", CategoryName: "Members"},
		{CategoryID: 20, ProjectID: 200, CategoryCode: "members", CategoryName: "Members"},
	}
	if matched := bestAutomaticFeatureMatch(features, []string{"members"}); matched != nil {
		t.Fatalf("feature code without project scope must be ambiguous: %#v", matched)
	}
}

func TestHierarchyCacheKeyIsolatesEnvironmentAndProject(t *testing.T) {
	projectA, projectB := 100, 200
	dev := hierarchyCacheKey(1, 10, &projectA, "", nil, "members")
	prod := hierarchyCacheKey(1, 20, &projectA, "", nil, "members")
	otherProject := hierarchyCacheKey(1, 10, &projectB, "", nil, "members")
	if dev == prod {
		t.Fatal("DEV and PROD must not share hierarchy cache keys")
	}
	if dev == otherProject {
		t.Fatal("Project A and Project B must not share hierarchy cache keys")
	}
}

func TestParentScopedHierarchySeparatesProjectsWithSameCodes(t *testing.T) {
	projectAFeatureID, projectBFeatureID := 11, 21
	projectA := []models.ProjectFeature{
		{CategoryID: projectAFeatureID, ProductID: 1, ProjectID: 100, CategoryCode: "BILL"},
		{CategoryID: 12, ProductID: 1, ProjectID: 100, ParentID: &projectAFeatureID, CategoryCode: "WATER"},
	}
	projectB := []models.ProjectFeature{
		{CategoryID: projectBFeatureID, ProductID: 1, ProjectID: 200, CategoryCode: "BILL"},
		{CategoryID: 22, ProductID: 1, ProjectID: 200, ParentID: &projectBFeatureID, CategoryCode: "WATER"},
	}

	rootA := findChildFeature(projectA, nil, "bill")
	rootB := findChildFeature(projectB, nil, "BILL")
	if rootA == nil || rootB == nil {
		t.Fatal("both scoped projects must resolve their own BILL feature")
	}
	waterA := findChildFeature(projectA, &rootA.CategoryID, "WATER")
	waterB := findChildFeature(projectB, &rootB.CategoryID, "water")
	if waterA == nil || waterA.CategoryID != 12 {
		t.Fatalf("PROJECT_A resolved incorrectly: %#v", waterA)
	}
	if waterB == nil || waterB.CategoryID != 22 {
		t.Fatalf("PROJECT_B resolved incorrectly: %#v", waterB)
	}
}

func TestParentScopedHierarchyDoesNotSelectSiblingOrCrossProduct(t *testing.T) {
	billID, otherID := 11, 13
	features := []models.ProjectFeature{
		{CategoryID: billID, ProductID: 1, ProjectID: 100, CategoryCode: "BILL"},
		{CategoryID: 12, ProductID: 1, ProjectID: 100, ParentID: &billID, CategoryCode: "WATER"},
		{CategoryID: otherID, ProductID: 1, ProjectID: 100, CategoryCode: "OTHER"},
		{CategoryID: 14, ProductID: 1, ProjectID: 100, ParentID: &otherID, CategoryCode: "WATER"},
	}
	if got := findChildFeature(features, &billID, "MISSING"); got != nil {
		t.Fatalf("missing child must not route elsewhere: %#v", got)
	}
	if got := findChildFeature(features, &billID, "WATER"); got == nil || got.CategoryID != 12 {
		t.Fatalf("parent-scoped lookup selected the wrong WATER: %#v", got)
	}
	if got := findFeatureByID(features, 999); got != nil {
		t.Fatalf("an ID outside the product/project cache must be rejected: %#v", got)
	}
}

func TestLegacyCategoryCodeBecomesAmbiguousAcrossParents(t *testing.T) {
	billID, otherID := 11, 13
	features := []models.ProjectFeature{
		{CategoryID: 12, ParentID: &billID, CategoryCode: "WATER"},
		{CategoryID: 14, ParentID: &otherID, CategoryCode: "WATER"},
	}
	if matches := findFeaturesByCode(features, "water", 2); len(matches) != 2 {
		t.Fatalf("legacy lookup must expose ambiguity instead of choosing a parent: %#v", matches)
	}
}

func TestFallbackToProductEnvironment(t *testing.T) {
	payload := map[string]any{
		"project_id":       1,
		"category_id":      2,
		"feature_path_ids": "2",
	}
	fallbackToProductEnvironment(payload, "not matched")

	if payload["routing_status"] != "UNCLASSIFIED" || payload["routing_method"] != "PRODUCT_ENVIRONMENT_DEFAULT" {
		t.Fatalf("routing metadata = %#v", payload)
	}
	if _, exists := payload["project_id"]; exists {
		t.Fatal("project_id must be removed on fallback")
	}
	if _, exists := payload["category_id"]; exists {
		t.Fatal("category_id must be removed on fallback")
	}
}

func TestFallbackToProjectPreservesProjectAndClearsCategory(t *testing.T) {
	payload := map[string]any{
		"project_id":       10,
		"category_id":      99,
		"feature_path_ids": "99",
	}
	fallbackToProject(payload, 10, "feature not found in scoped project")
	if payload["project_id"] != 10 {
		t.Fatalf("project scope was not preserved: %#v", payload)
	}
	if _, exists := payload["category_id"]; exists {
		t.Fatal("out-of-scope category must be cleared")
	}
	if payload["routing_method"] != "PROJECT_SCOPE_DEFAULT" {
		t.Fatalf("unexpected routing metadata: %#v", payload)
	}
}

func TestNormalizeCustomHierarchyExtractsMetadataActorProjectID(t *testing.T) {
	payload := map[string]any{
		"metadata": map[string]any{
			"actor": map[string]any{
				"id":         87,
				"project_id": 10,
			},
		},
	}

	normalizeCustomHierarchy(payload)

	if payload["project_id"] != 10 {
		t.Fatalf("expected payload project_id to be 10, got %#v", payload["project_id"])
	}
}

func TestNormalizeActorFieldPreservesProjectID(t *testing.T) {
	payload := map[string]any{
		"actor": map[string]any{
			"id":         87,
			"name":       "Admin",
			"role":       "admins",
			"project_id": 10,
		},
	}

	normalizeActorField(payload)

	actor, ok := payload["actor"].(map[string]any)
	if !ok {
		t.Fatalf("expected normalized actor map, got %#v", payload["actor"])
	}
	if actor["project_id"] != 10 {
		t.Fatalf("expected actor.project_id to be 10, got %#v", actor["project_id"])
	}
}

func TestFindDescendantFeatureMultiLevel(t *testing.T) {
	level1ID, level2ID := 10, 12
	pathIDsL1, pathIDsL2, pathIDsL3 := "10", "10,12", "10,12,15"
	fullPathL1, fullPathL2, fullPathL3 := "Finance", "Finance/Other Bills", "Finance/Other Bills/Add Batch"

	features := []models.ProjectFeature{
		{CategoryID: level1ID, ProductID: 1, ProjectID: 100, CategoryCode: "finance", CategoryName: "Finance", PathIDs: &pathIDsL1, FullPath: &fullPathL1, Level: 1},
		{CategoryID: level2ID, ProductID: 1, ProjectID: 100, ParentID: &level1ID, CategoryCode: "other_bills", CategoryName: "Other Bills", PathIDs: &pathIDsL2, FullPath: &fullPathL2, Level: 2},
		{CategoryID: 15, ProductID: 1, ProjectID: 100, ParentID: &level2ID, CategoryCode: "other_bills_add_batch", CategoryName: "Add Batch", PathIDs: &pathIDsL3, FullPath: &fullPathL3, Level: 3},
	}

	root := findDescendantFeature(features, nil, "finance")
	if root == nil || root.CategoryID != 10 {
		t.Fatalf("expected root Finance (10), got %#v", root)
	}

	// Resolve level 3 descendant under level 1 root
	descendant := findDescendantFeature(features, root, "other_bills_add_batch")
	if descendant == nil || descendant.CategoryID != 15 {
		t.Fatalf("expected descendant Add Batch (15), got %#v", descendant)
	}
	if *descendant.FullPath != "Finance/Other Bills/Add Batch" {
		t.Fatalf("expected FullPath Finance/Other Bills/Add Batch, got %s", *descendant.FullPath)
	}
}
func TestFindDescendantFeatureRejectsAmbiguousCodeAcrossParents(t *testing.T) {
	featureAID, featureBID := 10, 20
	features := []models.ProjectFeature{
		{CategoryID: featureAID, CategoryCode: "others_bill"},
		{CategoryID: featureBID, CategoryCode: "service"},
		{CategoryID: 11, ParentID: &featureAID, CategoryCode: "internet"},
		{CategoryID: 21, ParentID: &featureBID, CategoryCode: "internet"},
	}
	if got := findDescendantFeature(features, nil, "internet"); got != nil {
		t.Fatalf("ambiguous sub-feature must not select the first match: %#v", got)
	}
	if got := findDescendantFeature(features, &features[0], "internet"); got == nil || got.CategoryID != 11 {
		t.Fatalf("parent-scoped lookup should select category 11: %#v", got)
	}
}
