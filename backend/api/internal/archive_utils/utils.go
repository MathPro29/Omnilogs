package archive_utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

func parseDateFromIndexName(indexName string) (time.Time, error) {
	parts := strings.Split(indexName, "-")
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid index name structure: %s", indexName)
	}
	dateStr := parts[len(parts)-1]
	return time.Parse("2006.01.02", dateStr)
}

// Helper function to decode compressed payload to inspect/read stream (if needed)
func GetDecompressedReader(filePath string) (io.ReadCloser, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, err
	}
	return struct {
		io.Reader
		io.Closer
	}{
		Reader: gzipReader,
		Closer: gzipCloser{gzipReader, file},
	}, nil
}

type gzipCloser struct {
	gr *gzip.Reader
	f  *os.File
}

func (gc gzipCloser) Close() error {
	err1 := gc.gr.Close()
	err2 := gc.f.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
