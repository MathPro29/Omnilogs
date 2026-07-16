package usecase

import (
	"context"
	"fmt"
	"strings"

	"omnilogs-api/responses"
)

func (u *usecase) AuthorizeProductAccess(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64) error {
	return u.ensureProductAccess(ctx, actorUserID, platformAdmin, productID)
}

func (u *usecase) ensureProductAccess(ctx context.Context, actorUserID uint, platformAdmin bool, productID int64) error {
	if platformAdmin {
		return nil
	}
	hasAccess, err := u.repo.HasActiveProductMembership(ctx, actorUserID, productID)
	if err != nil {
		return err
	}
	if !hasAccess {
		return responses.ErrForbidden
	}
	return nil
}

func (u *usecase) resolveSearchIndices(ctx context.Context, productID int64) []string {
	indices := []string{fmt.Sprintf("omnilogs-product-%d-*", productID)}
	policies, err := u.repo.GetActiveIndexPolicies(ctx, productID)
	if err != nil {
		return indices
	}
	seen := map[string]struct{}{indices[0]: {}}
	for _, policy := range policies {
		prefix := strings.ToLower(strings.TrimSpace(policy.IndexPrefix))
		prefix = strings.ReplaceAll(prefix, "_", "-")
		prefix = strings.ReplaceAll(prefix, " ", "-")
		pattern := fmt.Sprintf("%s-*", prefix)
		if prefix == "" {
			continue
		}
		if _, exists := seen[pattern]; exists {
			continue
		}
		seen[pattern] = struct{}{}
		indices = append(indices, pattern)
	}
	return indices
}
