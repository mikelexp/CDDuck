package main

import (
	"path/filepath"
	"reflect"
	"testing"
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
