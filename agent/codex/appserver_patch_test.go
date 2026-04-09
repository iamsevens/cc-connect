package codex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chenhg5/cc-connect/core"
)

func TestAppServerTurnCompletedPatchesSessionSource(t *testing.T) {
	tmpDir := t.TempDir()
	codexHome := filepath.Join(tmpDir, ".codex")
	sessionsDir := filepath.Join(codexHome, "sessions")
	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		t.Fatalf("mkdir sessions dir: %v", err)
	}
	t.Setenv("CODEX_HOME", codexHome)

	sessionID := "appserver-thread-123"
	sessionFile := filepath.Join(sessionsDir, "rollout-"+sessionID+".jsonl")
	content := strings.Join([]string{
		`{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"` + sessionID + `","source":"exec","originator":"codex_exec","cwd":"C:/work/project"}}`,
		`{"timestamp":"2026-01-01T00:00:01Z","type":"response_item","payload":{"role":"user"}}`,
		"",
	}, "\n")
	if err := os.WriteFile(sessionFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}

	s := &appServerSession{
		events: make(chan core.Event, 4),
	}
	s.threadID.Store(sessionID)

	s.handleNotification("turn/completed", []byte(`{"threadId":"`+sessionID+`","turn":{"id":"turn-1","status":"completed"}}`))

	select {
	case event := <-s.events:
		if event.Type != core.EventResult {
			t.Fatalf("event.Type = %v, want %v", event.Type, core.EventResult)
		}
		if event.SessionID != sessionID {
			t.Fatalf("event.SessionID = %q, want %q", event.SessionID, sessionID)
		}
	default:
		t.Fatal("expected EventResult to be emitted")
	}

	data, err := os.ReadFile(sessionFile)
	if err != nil {
		t.Fatalf("read patched session file: %v", err)
	}
	firstLine := strings.SplitN(string(data), "\n", 2)[0]
	if !strings.Contains(firstLine, `"source":"cli"`) {
		t.Fatalf("expected patched source=cli, got %s", firstLine)
	}
	if !strings.Contains(firstLine, `"originator":"codex_cli_rs"`) {
		t.Fatalf("expected patched originator, got %s", firstLine)
	}
}
