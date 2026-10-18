package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

func newV7() string {
	return uuid.Must(uuid.NewV7()).String()
}

type RunSpec struct {
	Type string   `json:"type"`
	Args []string `json:"args"`
}

type Result struct {
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	Run         RunSpec   `json:"run"`
	Response    string    `json:"response"`
}

func writeResult(path string, r Result) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(r, "", "    ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}

func writeOutput(path string, w io.Writer, content string) error {
	if path == "" {
		_, err := io.WriteString(w, content)
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
