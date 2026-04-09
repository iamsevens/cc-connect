package codex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCodexSessionPath_WindowsSeparators(t *testing.T) {
	gotForward := normalizeCodexSessionPath(`D:/git/AURA-RAG`)
	gotBackward := normalizeCodexSessionPath(`D:\git\AURA-RAG`)
	if gotForward != gotBackward {
		t.Fatalf("windows path normalization mismatch: %q != %q", gotForward, gotBackward)
	}
}

func TestParseCodexSessionFile_MatchesNormalizedWindowsCwd(t *testing.T) {
	sessionFile := writeCodexSessionFile(t, `D:/git/AURA-RAG`)

	info := parseCodexSessionFile(sessionFile, normalizeCodexSessionPath(`D:\git\AURA-RAG`))
	if info == nil {
		t.Fatal("expected session to match normalized Windows cwd")
	}
	if info.ID != "session-1" {
		t.Fatalf("unexpected session ID: %q", info.ID)
	}
}

func TestParseCodexSessionFile_MatchesTrailingSlash(t *testing.T) {
	workDir := t.TempDir()
	sessionFile := writeCodexSessionFile(t, workDir+string(os.PathSeparator))

	info := parseCodexSessionFile(sessionFile, normalizeCodexSessionPath(workDir))
	if info == nil {
		t.Fatal("expected trailing separator cwd to match")
	}
}

func TestParseCodexSessionFile_FiltersDifferentCwd(t *testing.T) {
	sessionFile := writeCodexSessionFile(t, `D:/git/OTHER`)

	info := parseCodexSessionFile(sessionFile, normalizeCodexSessionPath(`D:\git\AURA-RAG`))
	if info != nil {
		t.Fatalf("expected different cwd to be filtered, got %+v", *info)
	}
}

func writeCodexSessionFile(t *testing.T, cwd string) string {
	t.Helper()

	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "session.jsonl")

	line, err := json.Marshal(struct {
		Type    string `json:"type"`
		Payload struct {
			ID  string `json:"id"`
			Cwd string `json:"cwd"`
		} `json:"payload"`
	}{
		Type: "session_meta",
		Payload: struct {
			ID  string `json:"id"`
			Cwd string `json:"cwd"`
		}{
			ID:  "session-1",
			Cwd: cwd,
		},
	})
	if err != nil {
		t.Fatalf("marshal session meta: %v", err)
	}

	if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
		t.Fatalf("write session file: %v", err)
	}
	return path
}
