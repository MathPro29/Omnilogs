package archive_utils

import (
	"strings"
	"testing"
)

func TestDecodeArchiveDataCSV(t *testing.T) {
	csvArchive := "log_id,timestamp,message,raw_payload\n" +
		`log-1,2026-01-01T00:00:00Z,hello,"{""product_id"":7,""payload"":{""message"":""hello""}}"` + "\n"

	data, err := decodeArchiveData(strings.NewReader(csvArchive), "logs.csv.gz")
	if err != nil {
		t.Fatalf("decode CSV archive: %v", err)
	}
	if len(data.Logs) != 1 || data.Logs[0]["_id"] != "log-1" {
		t.Fatalf("unexpected archive data: %#v", data)
	}
	source := data.Logs[0]["_source"].(map[string]any)
	if source["product_id"] != float64(7) {
		t.Fatalf("unexpected source: %#v", source)
	}
}

func TestDecodeArchiveDataRejectsInvalidCSVHeader(t *testing.T) {
	_, err := decodeArchiveData(strings.NewReader("id,payload\n1,{}\n"), "logs.csv.gz")
	if err == nil {
		t.Fatal("expected invalid header error")
	}
}
