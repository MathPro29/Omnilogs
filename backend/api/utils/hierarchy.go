package utils

import (
	"strconv"
	"strings"
)

// CategoryScopeAllowsParent keeps hierarchy navigation readable when a member
// is scoped to a child category. It never grants access to sibling branches.
func CategoryScopeAllowsParent(grantedPath *string, requestedCategoryID int, allowParentRead bool) bool {
	if !allowParentRead || grantedPath == nil || requestedCategoryID <= 0 {
		return false
	}
	needle := "," + strconv.Itoa(requestedCategoryID) + ","
	return strings.Contains(","+*grantedPath+",", needle)
}
