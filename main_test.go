package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTruncateWithEllipsis(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  string
	}{
		{name: "room for ellipsis", text: "directory", width: 5, want: "dire…"},
		{name: "single column", text: "directory", width: 1, want: "…"},
		{name: "unicode", text: "café-directory", width: 5, want: "café…"},
		{name: "zero width", text: "directory", width: 0, want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateWithEllipsis(tt.text, tt.width); got != tt.want {
				t.Errorf("truncateWithEllipsis(%q, %d) = %q, want %q", tt.text, tt.width, got, tt.want)
			}
		})
	}
}

func TestPromoteHistoryMovesSelectionToFrontAndDeduplicates(t *testing.T) {
	got := promoteHistory([]string{"/work", "/home", "/work", "/tmp"}, "/home")
	want := []string{"/home", "/work", "/tmp"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("promoteHistory() = %#v, want %#v", got, want)
	}
}

func TestHistoryFiltering(t *testing.T) {
	m := Model{
		history:       []string{"/home/mikele/project", "/tmp"},
		historyFilter: "proj",
	}
	got := m.filteredHistory()
	want := []string{"/home/mikele/project"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filteredHistory() = %#v, want %#v", got, want)
	}
}

func TestSaveAndLoadHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cdduck", "history.json")
	want := []string{"/home/mikele/project", "/tmp"}
	if err := saveHistory(path, want); err != nil {
		t.Fatalf("saveHistory() error = %v", err)
	}
	got, err := loadHistory(path)
	if err != nil {
		t.Fatalf("loadHistory() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loadHistory() = %#v, want %#v", got, want)
	}
}

func TestParseArgsOpensHistory(t *testing.T) {
	dir, showVersion, openHistory, err := parseArgs([]string{"-H"})
	if err != nil {
		t.Fatalf("parseArgs() error = %v", err)
	}
	if showVersion {
		t.Error("showVersion = true, want false")
	}
	if !openHistory {
		t.Error("openHistory = false, want true")
	}
	if dir == "" {
		t.Error("dir is empty, want current directory")
	}
}

func TestOpeningHistoryEntryNavigatesMainWindow(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	if err := os.Mkdir(target, 0700); err != nil {
		t.Fatalf("Mkdir() error = %v", err)
	}
	m := Model{
		currentDir:  root,
		history:     []string{target},
		historyOpen: true,
	}
	m.loadItems()

	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Fatalf("Update() command = %v, want nil", cmd)
	}
	got := updated.(Model)
	if got.currentDir != target {
		t.Errorf("currentDir = %q, want %q", got.currentDir, target)
	}
	if got.historyOpen {
		t.Error("history modal remained open after choosing an entry")
	}
	if got.selected != "" {
		t.Errorf("selected = %q, want empty (navigation should not exit)", got.selected)
	}
}
