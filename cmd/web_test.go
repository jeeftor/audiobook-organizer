package cmd

import (
	"strings"
	"testing"
)

func TestWebBindWarning(t *testing.T) {
	tests := []struct {
		host        string
		inContainer bool
		want        string
	}{
		{host: "127.0.0.1", inContainer: true, want: "published Docker ports cannot reach it"},
		{host: "0.0.0.0", want: "network-reachable"},
		{host: "127.0.0.1", want: ""},
	}
	for _, tt := range tests {
		if got := webBindWarning(tt.host, tt.inContainer); tt.want != "" &&
			!strings.Contains(got, tt.want) {
			t.Errorf("webBindWarning(%q, %t) = %q, want %q", tt.host, tt.inContainer, got, tt.want)
		} else if tt.want == "" && got != "" {
			t.Errorf("webBindWarning(%q, %t) = %q, want empty", tt.host, tt.inContainer, got)
		}
	}
}
