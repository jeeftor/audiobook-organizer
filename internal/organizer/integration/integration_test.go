//go:build integration

package integration

// This file ensures the integration package is recognized as a test package
// and provides a place for package-level documentation.

// The integration package contains end-to-end tests that verify the behavior
// of the organizer package against real files and directories.
//
// These tests are slower than unit tests and require filesystem access.
// Use `go test -tags=integration` to run them.
