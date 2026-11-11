package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/eugenioenko/ttt/internal/config"
)

func TestReloadSettingsCommand(t *testing.T) {
	h := newTestHarness(t, 80, 24)
	defer h.stop()

	// Verify defaults: line numbers on, tab size 4, gutter style compact
	if !h.app.EditorGroup.LineNumbers {
		t.Fatal("expected line numbers enabled by default")
	}
	if h.app.EditorGroup.Editor.LineNumbers != true {
		t.Fatal("expected editor pane line numbers enabled by default")
	}
	if h.app.EditorGroup.TabSize != 4 {
		t.Fatal("expected tab size 4 by default")
	}
	if h.app.EditorGroup.GutterStyle != "compact" {
		t.Fatalf("expected gutter style 'compact' by default, got %q", h.app.EditorGroup.GutterStyle)
	}

	// Create modified settings and apply them via ApplySettings
	s := config.DefaultSettings()
	s.Editor.LineNumbers = false
	s.Editor.TabSize = 2
	s.Editor.GutterStyle = "minimal"
	s.Editor.InsertFinalNewline = false
	s.Editor.TrimTrailingWhitespace = true
	s.Search.Debounce = 500

	h.app.ApplySettings(s)
	h.redraw()

	// Verify editor group settings updated
	if h.app.EditorGroup.LineNumbers {
		t.Error("expected line numbers disabled after apply")
	}
	if h.app.EditorGroup.Editor.LineNumbers {
		t.Error("expected editor pane line numbers disabled after apply")
	}
	if h.app.EditorGroup.TabSize != 2 {
		t.Errorf("expected tab size 2 after apply, got %d", h.app.EditorGroup.TabSize)
	}
	if h.app.EditorGroup.Editor.TabSize != 2 {
		t.Errorf("expected editor pane tab size 2 after apply, got %d", h.app.EditorGroup.Editor.TabSize)
	}
	if h.app.EditorGroup.GutterStyle != "minimal" {
		t.Errorf("expected gutter style 'minimal' after apply, got %q", h.app.EditorGroup.GutterStyle)
	}
	if h.app.EditorGroup.Editor.GutterStyle != "minimal" {
		t.Errorf("expected editor pane gutter style 'minimal' after apply, got %q", h.app.EditorGroup.Editor.GutterStyle)
	}
	if h.app.EditorGroup.InsertFinalNewline {
		t.Error("expected InsertFinalNewline false after apply")
	}
	if !h.app.EditorGroup.TrimTrailingWhitespace {
		t.Error("expected TrimTrailingWhitespace true after apply")
	}
	if h.app.Search.Debounce.DelayMs != 500 {
		t.Errorf("expected search debounce 500 after apply, got %d", h.app.Search.Debounce.DelayMs)
	}

	// Verify the settings pointer was updated
	if h.app.Settings.Editor.LineNumbers {
		t.Error("expected Settings.Editor.LineNumbers false after apply")
	}
	if h.app.Settings.Editor.TabSize != 2 {
		t.Errorf("expected Settings.Editor.TabSize 2 after apply, got %d", h.app.Settings.Editor.TabSize)
	}
}

func TestReloadSettingsCommandRegistered(t *testing.T) {
	h := newTestHarness(t, 80, 24)
	defer h.stop()

	// Verify command is registered
	cmd, ok := h.reg.Get("settings.reload")
	if !ok {
		t.Fatal("expected settings.reload command to be registered")
	}
	if cmd.Title != "Reload Settings" {
		t.Errorf("expected title 'Reload Settings', got %q", cmd.Title)
	}
}

func TestReloadSettingsPreviewTabs(t *testing.T) {
	h := newTestHarness(t, 80, 24)
	defer h.stop()

	if !h.app.EditorGroup.PreviewTabs {
		t.Fatal("expected preview tabs enabled on the editor group at startup")
	}

	a := filepath.Join(h.dir, "alpha.txt")
	b := filepath.Join(h.dir, "beta.txt")
	c := filepath.Join(h.dir, "gamma.txt")
	h.app.EditorGroup.OpenFile(a)
	h.app.EditorGroup.OpenFile(b) // preview enabled: b replaces a
	h.redraw()
	if got := h.app.EditorGroup.TabCount(); got != 1 {
		t.Fatalf("with preview enabled: tabs = %d, want 1", got)
	}

	// Live-apply disabled preview tabs; the on-disk settings still default to
	// enabled, so tab behavior must follow the applied value, not the file.
	s := config.DefaultSettings()
	disabled := false
	s.Editor.PreviewTabs = &disabled
	h.app.ApplySettings(s)
	h.redraw()
	if h.app.EditorGroup.PreviewTabs {
		t.Fatal("expected PreviewTabs false on the editor group after ApplySettings")
	}
	h.app.EditorGroup.OpenFile(c) // must not replace the open preview tab
	h.redraw()
	if got := h.app.EditorGroup.TabCount(); got != 2 {
		t.Fatalf("after disabling preview tabs: tabs = %d, want 2", got)
	}
	if got := h.app.EditorGroup.ActiveFilePath(); got != c {
		t.Fatalf("active tab = %q, want %q", got, c)
	}

	// Reload from disk with preview tabs disabled changes tab-opening behavior.
	cfgDir := filepath.Join(h.dir, "config")
	raw := `{"editor":{"previewTabs":false}}`
	if err := os.WriteFile(filepath.Join(cfgDir, "settings.json"), []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	h.exec("settings.reload")
	h.redraw()
	if h.app.EditorGroup.PreviewTabs {
		t.Fatal("expected PreviewTabs false after settings.reload")
	}
	h.app.EditorGroup.OpenFile(a) // active tab c is not a preview: a appends
	h.redraw()
	if got := h.app.EditorGroup.TabCount(); got != 3 {
		t.Fatalf("after reload with preview disabled: tabs = %d, want 3", got)
	}
}

func TestApplySettingsRestoresDefaults(t *testing.T) {
	h := newTestHarness(t, 80, 24)
	defer h.stop()

	// Modify settings to non-default values
	modified := config.DefaultSettings()
	modified.Editor.LineNumbers = false
	modified.Editor.TabSize = 8
	modified.Editor.GutterStyle = "extended"
	h.app.ApplySettings(modified)

	if h.app.EditorGroup.LineNumbers {
		t.Error("expected line numbers disabled after first apply")
	}

	// Apply defaults back and verify restoration
	defaults := config.DefaultSettings()
	h.app.ApplySettings(defaults)

	if !h.app.EditorGroup.LineNumbers {
		t.Error("expected line numbers re-enabled after applying defaults")
	}
	if h.app.EditorGroup.TabSize != 4 {
		t.Errorf("expected tab size 4 after applying defaults, got %d", h.app.EditorGroup.TabSize)
	}
	if h.app.EditorGroup.GutterStyle != "compact" {
		t.Errorf("expected gutter style 'compact' after applying defaults, got %q", h.app.EditorGroup.GutterStyle)
	}
}
