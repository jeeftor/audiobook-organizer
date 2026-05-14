package main

import (
	"os"
	"testing"

	"github.com/jeeftor/audiobook-organizer/cmd"
)

func TestMainFlags(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "missing_required_flag",
			args:    []string{"audiobook-organizer"},
			wantErr: true,
		},
		{
			name:    "valid_flags",
			args:    []string{"audiobook-organizer", "--dir=.", "--dry-run"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			err := cmd.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("cmd.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
