package repository

import (
	"errors"
	"testing"
)

func TestTranslateActivePolicyConstraint(t *testing.T) {
	err := translatePolicyWriteError(errors.New(`ERROR: duplicate key value violates unique constraint "uq_retention_policy_product_environment_active" (SQLSTATE 23505)`))
	if !errors.Is(err, ErrActivePolicyExists) {
		t.Fatalf("constraint error was not translated: %v", err)
	}
}
