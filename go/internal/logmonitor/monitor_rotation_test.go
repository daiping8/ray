// Copyright 2026 The Ray Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logmonitor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReopenIfNecessaryAfterReplacementParity(t *testing.T) {
	logsDir := t.TempDir()
	logPath := createLogFile(t, logsDir, "raylet.err", "line1\n")

	info := &logFileInfo{filename: logPath}
	if err := info.openAtCurrentPosition(); err != nil {
		t.Fatalf("openAtCurrentPosition() error = %v", err)
	}
	info.filePosition = int64(len("line1\nline2\n"))

	replacementPath := filepath.Join(logsDir, "replacement.err")
	if err := os.WriteFile(replacementPath, []byte("line2\n"), 0o644); err != nil {
		t.Fatalf("write replacement: %v", err)
	}
	if err := os.Rename(replacementPath, logPath); err != nil {
		t.Fatalf("replace file: %v", err)
	}

	if err := info.reopenIfNecessary(); err != nil {
		t.Fatalf("reopenIfNecessary() error = %v", err)
	}
	if info.filePosition != 0 {
		t.Fatalf("filePosition = %d, want 0 after replacement with smaller file", info.filePosition)
	}
}

func TestReopenIfNecessaryAfterTruncationParity(t *testing.T) {
	logsDir := t.TempDir()
	logPath := createLogFile(t, logsDir, "raylet.err", "line1\nline2\n")

	info := &logFileInfo{filename: logPath}
	if err := info.openAtCurrentPosition(); err != nil {
		t.Fatalf("openAtCurrentPosition() error = %v", err)
	}
	info.filePosition = int64(len("line1\nline2\n"))

	if err := os.WriteFile(logPath, []byte("short\n"), 0o644); err != nil {
		t.Fatalf("truncate write: %v", err)
	}

	if err := info.reopenIfNecessary(); err != nil {
		t.Fatalf("reopenIfNecessary() error = %v", err)
	}
	if info.filePosition != 0 {
		t.Fatalf("filePosition = %d, want 0 after truncation", info.filePosition)
	}
	line := make([]byte, len("short\n"))
	if _, err := info.fileHandle.Read(line); err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if got := string(line); got != "short\n" {
		t.Fatalf("first line after truncation = %q, want %q", got, "short\n")
	}
}

func TestReopenIfNecessaryKeepsPositionForLargerReplacementParity(t *testing.T) {
	logsDir := t.TempDir()
	logPath := createLogFile(t, logsDir, "raylet.err", "line1\nline2\n")

	info := &logFileInfo{filename: logPath}
	if err := info.openAtCurrentPosition(); err != nil {
		t.Fatalf("openAtCurrentPosition() error = %v", err)
	}
	info.filePosition = int64(len("line1\n"))

	replacementPath := filepath.Join(logsDir, "replacement.err")
	if err := os.WriteFile(replacementPath, []byte("line1\nline2\nline3\n"), 0o644); err != nil {
		t.Fatalf("write replacement: %v", err)
	}
	if err := os.Rename(replacementPath, logPath); err != nil {
		t.Fatalf("replace file: %v", err)
	}

	if err := info.reopenIfNecessary(); err != nil {
		t.Fatalf("reopenIfNecessary() error = %v", err)
	}
	if info.filePosition != int64(len("line1\n")) {
		t.Fatalf("filePosition = %d, want %d after larger replacement", info.filePosition, len("line1\n"))
	}
}
