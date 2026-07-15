package main

import (
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	rows, err := db.Raw("SELECT masking_rule_id, product_id, field_key, field_path, match_type, mask_type, mask_value, secret_handling_mode, encryption_required, view_requires_password, is_active FROM log_masking_rules").Rows()
	if err != nil {
		log.Fatalf("failed to query raw: %v", err)
	}
	defer rows.Close()

	fmt.Println("=== Masking Rules Raw ===")
	for rows.Next() {
		var masking_rule_id int
		var product_id *int
		var field_key *string
		var field_path *string
		var match_type *string
		var mask_type *string
		var mask_value *string
		var secret_handling_mode string
		var encryption_required bool
		var view_requires_password bool
		var is_active bool

		err = rows.Scan(&masking_rule_id, &product_id, &field_key, &field_path, &match_type, &mask_type, &mask_value, &secret_handling_mode, &encryption_required, &view_requires_password, &is_active)
		if err != nil {
			log.Fatalf("scan error: %v", err)
		}

		pidVal := -1
		if product_id != nil {
			pidVal = *product_id
		}
		keyVal := "nil"
		if field_key != nil {
			keyVal = *field_key
		}
		pathVal := "nil"
		if field_path != nil {
			pathVal = *field_path
		}
		maskVal := "nil"
		if mask_value != nil {
			maskVal = *mask_value
		}
		fmt.Printf("RuleID: %d, ProductID: %d, Key: %s, Path: %s, MaskValue: %s, SecretMode: %s, EncryptionRequired: %v, IsActive: %v\n",
			masking_rule_id, pidVal, keyVal, pathVal, maskVal, secret_handling_mode, encryption_required, is_active)
	}
}
