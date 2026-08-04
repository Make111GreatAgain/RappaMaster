package test

import (
	"BHLayer2Node/paradigm"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogWriter tests the LogWriter functionality
func TestLogWriter(t *testing.T) {
	logPath := t.TempDir()
	debug := false

	// Create and initialize LogWriter
	logWriter := paradigm.NewLogWriter(logPath, debug)
	if err := logWriter.Init(); err != nil {
		t.Fatalf("Failed to initialize LogWriter: %v", err)
	}
	defer logWriter.Close()

	// Log messages of different levels
	logWriter.Log("DEBUG", "This is a debug message for testing")
	logWriter.Log("INFO", "This is an info message for testing")
	logWriter.Log("WARNING", "This is a warning message for testing")
	logWriter.Log("ERROR", "This is an error message for testing")
	logWriter.Log("CRITICAL", "This is a critical message for testing")
	logWriter.Log("NETWORK", "This is a network message for testing")
	logWriter.Log("SCHEDULE", "This is an schedule message for testing")
	logWriter.Log("CHAINUP", "This is a chainup message for testing")

	// Verify log file is created
	files, err := os.ReadDir(logPath)
	if err != nil {
		t.Fatalf("Failed to read log directory: %v", err)
	}
	if len(files) == 0 {
		t.Fatalf("No log file was created in the directory: %s", logPath)
	}

	var contents strings.Builder
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(logPath, file.Name()))
		if err != nil {
			t.Fatalf("Failed to read log file %s: %v", file.Name(), err)
		}
		contents.Write(content)
	}

	expectedMessages := []string{
		"This is a debug message for testing",
		"This is an info message for testing",
		"This is a warning message for testing",
		"This is an error message for testing",
		"This is a critical message for testing",
		"This is a network message for testing",
		"This is an schedule message for testing",
		"This is a chainup message for testing",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(contents.String(), msg) {
			t.Errorf("Log file does not contain expected message: %s", msg)
		}
	}

	//// Cleanup
	//if err := os.RemoveAll(logPath); err != nil {
	//	t.Logf("Failed to clean up log directory: %v", err)
	//}
}
