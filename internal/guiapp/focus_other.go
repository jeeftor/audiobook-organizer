//go:build gui && !darwin

package guiapp

// activateForDialog is a no-op on non-macOS platforms.
func activateForDialog() {}

// selectDirectoryNative returns "" on non-macOS; caller falls back to Wails runtime dialog.
func selectDirectoryNative(title string) string { return "" }

// EnableDevTools is a no-op on non-macOS platforms.
func EnableDevTools() {}

// PatchDevToolsAfterInit is a no-op on non-macOS platforms.
func PatchDevToolsAfterInit() {}

// OpenWebInspector is a no-op on non-macOS platforms.
func OpenWebInspector() {}
