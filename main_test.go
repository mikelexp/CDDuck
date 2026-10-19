package main

import "testing"

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
