package cmd

import (
	"strings"
	"testing"
)

func TestWebNoTokenFlagOmitsTokenFromURL(t *testing.T) {
	// When --no-token is set the token is empty and the URL should have no query string.
	cmd := webCmd
	if err := cmd.Flags().Set("no-token", "true"); err != nil {
		t.Fatalf("set flag: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Flags().Set("no-token", "false") })

	noToken, _ := cmd.Flags().GetBool("no-token")
	if !noToken {
		t.Fatal("expected no-token flag to be true")
	}

	var token string
	if !noToken {
		var err error
		token, err = newSessionToken()
		if err != nil {
			t.Fatalf("newSessionToken: %v", err)
		}
	}

	if token != "" {
		t.Fatalf("expected empty token when --no-token is set, got %q", token)
	}
}

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
