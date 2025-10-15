package json

import (
	"encoding/json"
	"fmt"
	"os"
)

// Reader handles file reading operations
type Reader struct{}

// NewReader creates a new filesystem reader
func NewReader() *Reader {
	return &Reader{}
}

// ReadJSON reads and unmarshals JSON from a file
func (r *Reader) ReadJSON(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return nil
}
