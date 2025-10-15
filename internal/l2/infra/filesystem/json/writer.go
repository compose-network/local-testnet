package json

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Writer handles file writing operations
type Writer struct{}

// NewWriter creates a new filesystem writer
func NewWriter() *Writer {
	return &Writer{}
}

// WriteJSON writes data as JSON to the specified path
func (w *Writer) WriteJSON(path string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if err := w.ensureDir(path); err != nil {
		return err
	}

	if err := os.WriteFile(path, append(content, '\n'), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// WriteBytes writes raw bytes to the specified path
func (w *Writer) WriteBytes(path string, data []byte) error {
	if err := w.ensureDir(path); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ensureDir ensures the parent directory of a file exists
func (w *Writer) ensureDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	return nil
}
