package usecase

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"omnilogs-api/models"
)

// Hierarchy matching is kept separate from database loading and payload
// orchestration so the ingestion flow remains easy to follow and test.

func findFeatureByID(features []models.ProjectFeature, id int) *models.ProjectFeature {
	for i := range features {
		if features[i].CategoryID == id {
			return &features[i]
		}
	}
	return nil
}

func splitHierarchyPath(val string) []string {
	val = strings.ReplaceAll(val, ">", "/")
	parts := strings.Split(val, "/")
	res := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			res = append(res, p)
		}
	}
	return res
}

func formatTitleCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.ReplaceAll(s, "-", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
		}
	}
	return strings.Join(words, " ")
}

func matchFeature(f models.ProjectFeature, code string) bool {
	code = strings.TrimSpace(code)
	if code == "" {
		return false
	}
	if strings.EqualFold(strings.TrimSpace(f.CategoryCode), code) || strings.EqualFold(strings.TrimSpace(f.CategoryName), code) {
		return true
	}
	cSlug := compactHierarchyValue(f.CategoryCode)
	nSlug := compactHierarchyValue(f.CategoryName)
	hSlug := compactHierarchyValue(code)
	return hSlug != "" && (cSlug == hSlug || nSlug == hSlug)
}

func findDescendantFeature(features []models.ProjectFeature, parent *models.ProjectFeature, code string) *models.ProjectFeature {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil
	}
	for i := range features {
		if !matchFeature(features[i], code) {
			continue
		}
		if parent == nil && features[i].ParentID == nil {
			return &features[i]
		}
		if parent != nil && features[i].ParentID != nil && *features[i].ParentID == parent.CategoryID {
			return &features[i]
		}
	}
	if parent != nil && parent.PathIDs != nil {
		parentIDStr := fmt.Sprint(parent.CategoryID)
		for i := range features {
			if !matchFeature(features[i], code) || features[i].PathIDs == nil {
				continue
			}
			for _, id := range strings.Split(*features[i].PathIDs, ",") {
				if id == parentIDStr && features[i].CategoryID != parent.CategoryID {
					return &features[i]
				}
			}
		}
	}
	if parent == nil {
		var match *models.ProjectFeature
		for i := range features {
			if !matchFeature(features[i], code) {
				continue
			}
			if match != nil {
				return nil
			}
			match = &features[i]
		}
		return match
	}
	return nil
}

func findFeaturesByCode(features []models.ProjectFeature, code string, limit int) []models.ProjectFeature {
	matches := make([]models.ProjectFeature, 0, limit)
	for i := range features {
		if strings.EqualFold(strings.TrimSpace(features[i].CategoryCode), strings.TrimSpace(code)) {
			matches = append(matches, features[i])
			if len(matches) == limit {
				break
			}
		}
	}
	return matches
}

var hierarchyTokenBoundary = regexp.MustCompile("[^a-zA-Z0-9]+")

// hierarchyRouteHints only considers stable route/hierarchy fields and never
// arbitrary request-body values.
func hierarchyRouteHints(payload map[string]any, featureCode string) []string {
	values := []string{featureCode, firstString(payload, "route_key", "feature_name", "event_name", "request_path", "route_pattern")}
	for _, key := range []string{"request_path", "route_pattern", "feature_name", "event_name"} {
		values = append(values, firstString(payload, key))
	}
	if action, ok := payload["action"].(map[string]any); ok {
		values = append(values, firstString(action, "operation", "resource_path", "route"))
	}
	if metadata, ok := payload["metadata"].(map[string]any); ok {
		values = append(values, firstString(metadata, "featureCode", "subFeatureCode", "feature_code", "sub_feature_code"))
		if action, ok := metadata["action"].(map[string]any); ok {
			values = append(values, firstString(action, "operation", "resource_path", "route"))
		}
	}
	if routing, ok := payload["routing"].(map[string]any); ok {
		if dimensions, ok := routing["dimensions"].(map[string]any); ok {
			values = append(values, firstString(dimensions, "featureCode", "subFeatureCode", "feature_code", "sub_feature_code", "operation"))
		}
	}
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func bestAutomaticFeatureMatch(features []models.ProjectFeature, hints []string) *models.ProjectFeature {
	bestScore, bestIndex, tied := 0, -1, false
	for i := range features {
		score := automaticFeatureScore(features[i], hints)
		if score > bestScore {
			bestScore, bestIndex, tied = score, i, false
		} else if score > 0 && score == bestScore {
			tied = true
		}
	}
	if bestIndex < 0 || bestScore < 70 || tied {
		return nil
	}
	return &features[bestIndex]
}

func automaticFeatureScore(feature models.ProjectFeature, hints []string) int {
	candidateValues := []string{feature.CategoryCode, feature.CategoryName}
	best := 0
	for _, candidate := range candidateValues {
		candidateCompact := compactHierarchyValue(candidate)
		candidateTokens := hierarchyTokens(candidate)
		if candidateCompact == "" || len(candidateCompact) < 4 {
			continue
		}
		for _, hint := range hints {
			hintCompact := compactHierarchyValue(hint)
			hintTokens := hierarchyTokens(hint)
			score := 0
			switch {
			case candidateCompact == hintCompact:
				score = 100
			case strings.Contains(hintCompact, candidateCompact):
				score = 90
			case tokenSubset(candidateTokens, hintTokens):
				score = 80 + minInt(len(candidateTokens), 9)
			case len(hintCompact) >= 5 && levenshteinDistance(candidateCompact, hintCompact) <= 2:
				score = 70
			}
			if score > best {
				best = score
			}
		}
	}
	return best
}

func hierarchyTokens(value string) []string {
	var expanded strings.Builder
	for i, r := range value {
		if i > 0 && unicode.IsUpper(r) {
			expanded.WriteByte(' ')
		}
		expanded.WriteRune(unicode.ToLower(r))
	}
	parts := hierarchyTokenBoundary.Split(expanded.String(), -1)
	tokens := make([]string, 0, len(parts))
	for _, part := range parts {
		part = singularHierarchyToken(strings.TrimSpace(part))
		if part != "" {
			tokens = append(tokens, part)
		}
	}
	sort.Strings(tokens)
	return tokens
}

func compactHierarchyValue(value string) string { return strings.Join(hierarchyTokens(value), "") }

func singularHierarchyToken(value string) string {
	if len(value) > 4 && strings.HasSuffix(value, "ies") {
		return strings.TrimSuffix(value, "ies") + "y"
	}
	if len(value) > 3 && strings.HasSuffix(value, "s") && !strings.HasSuffix(value, "ss") {
		return strings.TrimSuffix(value, "s")
	}
	return value
}

func tokenSubset(candidate, hint []string) bool {
	if len(candidate) == 0 || len(hint) == 0 {
		return false
	}
	available := make(map[string]struct{}, len(hint))
	for _, token := range hint {
		available[token] = struct{}{}
	}
	for _, token := range candidate {
		if _, exists := available[token]; !exists {
			return false
		}
	}
	return true
}

func levenshteinDistance(left, right string) int {
	if left == right {
		return 0
	}
	rightRunes := []rune(right)
	previous := make([]int, len(rightRunes)+1)
	for j := range previous {
		previous[j] = j
	}
	for i, leftRune := range []rune(left) {
		current := make([]int, len(rightRunes)+1)
		current[0] = i + 1
		for j, rightRune := range rightRunes {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[j+1] = minInt(minInt(current[j]+1, previous[j+1]+1), previous[j]+cost)
		}
		previous = current
	}
	return previous[len(previous)-1]
}

func findChildFeature(features []models.ProjectFeature, parentID *int, code string) *models.ProjectFeature {
	for i := range features {
		if !strings.EqualFold(strings.TrimSpace(features[i].CategoryCode), strings.TrimSpace(code)) {
			continue
		}
		if (parentID == nil) != (features[i].ParentID == nil) {
			continue
		}
		if parentID == nil || *parentID == *features[i].ParentID {
			return &features[i]
		}
	}
	return nil
}
