package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"tetris/internal/engine"
)

func TestLearnRejectsHistoricalIncompleteAndInvalidLogs(t *testing.T) {
	for _, data := range []string{"{bad json}\n", "{\"session_id\":\"old\",\"had_ai_plan\":true,\"ai_plan_matched\":true}\n"} {
		path := filepath.Join(t.TempDir(), "source.jsonl")
		out := filepath.Join(t.TempDir(), "model.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		if err := run(path, out, 10, engine.RiskTrainingConfig{}); err == nil {
			t.Fatal("invalid/historical source accepted")
		}
		if _, err := os.Stat(out); !os.IsNotExist(err) {
			t.Fatal("invalid training published model")
		}
	}
	if err := run("missing", "unused", 1, engine.RiskTrainingConfig{}); err == nil || !strings.Contains(err.Error(), "decisions") {
		t.Fatal("invalid sample limit accepted")
	}
}
