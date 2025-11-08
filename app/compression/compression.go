package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strings"
)

// Compressor handles content compression
type Compressor struct{}

// NewCompressor creates a new compressor
func NewCompressor() *Compressor {
	return &Compressor{}
}

// SupportsGzip checks if the Accept-Encoding header supports gzip
func (c *Compressor) SupportsGzip(acceptEncoding string) bool {
	if acceptEncoding == "" {
		return false
	}

	encodings := strings.Split(acceptEncoding, ",")
	for _, enc := range encodings {
		if strings.TrimSpace(enc) == "gzip" {
			return true
		}
	}
	return false
}

// CompressGzip compresses data using gzip
func (c *Compressor) CompressGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write compressed data: %w", err)
	}

	if err := gzWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}
