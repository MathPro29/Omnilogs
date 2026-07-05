package main

import (
	"fmt"
	"log"
	"os"

	"omnilogs-api/configs"
	"omnilogs-api/models"
)

func main() {
	os.Setenv("ENV_PATH", "../.env")
	env := configs.LoadEnv()
	db, err := configs.ConnectDB(env)
	if err != nil {
		log.Fatalf("failed to connect db: %v", err)
	}

	var archives []models.LogArchive
	db.Find(&archives)
	fmt.Println("=== Log Archives ===")
	for _, a := range archives {
		statusStr := "nil"
		if a.Status != nil {
			statusStr = *a.Status
		}
		totalLogsVal := 0
		if a.TotalLogs != nil {
			totalLogsVal = *a.TotalLogs
		}
		sizeVal := int64(0)
		if a.CompressedSizeBytes != nil {
			sizeVal = *a.CompressedSizeBytes
		}
		fmt.Printf("ArchiveID: %s, FilePath: %s, Status: %s, TotalLogs: %d, CompressedSize: %d\n", a.ArchiveID, *a.FilePath, statusStr, totalLogsVal, sizeVal)
	}
}
