package pgdump

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const (
	// Magic bytes for PostgreSQL dump format
	PGDumpMagic = "PGDMP"

	// Default PostgreSQL dump format version
	DefaultVersion = "1.14"
)

// Header represents the PostgreSQL dump file header
type Header struct {
	Version      string
	DatabaseName string
	Encoding     string
	Timestamp    time.Time
}

// NewHeader creates a new dump header with default values
func NewHeader() *Header {
	return &Header{
		Version:   DefaultVersion,
		Timestamp: time.Now(),
		Encoding:  "UTF8",
	}
}

// Write writes the header to the given writer
func (h *Header) Write(w io.Writer) error {
	// Write magic bytes
	if _, err := w.Write([]byte(PGDumpMagic)); err != nil {
		return fmt.Errorf("failed to write magic bytes: %w", err)
	}

	// Write version as major.minor (simplified format)
	// In real pg_dump, this is more complex
	versionBytes := []byte{1, 14, 0} // v1.14.0
	if _, err := w.Write(versionBytes); err != nil {
		return fmt.Errorf("failed to write version: %w", err)
	}

	// Write timestamp (Unix timestamp as int64)
	timestamp := h.Timestamp.Unix()
	if err := binary.Write(w, binary.LittleEndian, timestamp); err != nil {
		return fmt.Errorf("failed to write timestamp: %w", err)
	}

	// Write database name length and name
	dbNameBytes := []byte(h.DatabaseName)
	nameLen := uint32(len(dbNameBytes))
	if err := binary.Write(w, binary.LittleEndian, nameLen); err != nil {
		return fmt.Errorf("failed to write database name length: %w", err)
	}
	if _, err := w.Write(dbNameBytes); err != nil {
		return fmt.Errorf("failed to write database name: %w", err)
	}

	// Write encoding length and encoding
	encodingBytes := []byte(h.Encoding)
	encodingLen := uint32(len(encodingBytes))
	if err := binary.Write(w, binary.LittleEndian, encodingLen); err != nil {
		return fmt.Errorf("failed to write encoding length: %w", err)
	}
	if _, err := w.Write(encodingBytes); err != nil {
		return fmt.Errorf("failed to write encoding: %w", err)
	}

	return nil
}