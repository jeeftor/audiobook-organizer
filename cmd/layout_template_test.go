package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommandIncludesLayoutTemplateFlag(t *testing.T) {
	flag := rootCmd.Flags().Lookup("layout-template")
	if flag == nil {
		t.Fatal("root command missing layout-template flag")
	}
	if flag.DefValue != "" {
		t.Errorf("layout-template default = %q, want empty", flag.DefValue)
	}

	aliases := envAliases["layout-template"]
	if len(aliases) != 2 {
		t.Fatalf("layout-template aliases = %v, want 2 aliases", aliases)
	}
	if aliases[0] != "AO_LAYOUT_TEMPLATE" {
		t.Errorf("first layout-template alias = %q, want AO_LAYOUT_TEMPLATE", aliases[0])
	}
	if aliases[1] != "AUDIOBOOK_ORGANIZER_LAYOUT_TEMPLATE" {
		t.Errorf(
			"second layout-template alias = %q, want AUDIOBOOK_ORGANIZER_LAYOUT_TEMPLATE",
			aliases[1],
		)
	}
}

func TestRootCommandIncludesLayoutTemplateHelpCommand(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"layout-template"})
	if err != nil {
		t.Fatalf("Find(layout-template) error = %v", err)
	}
	if cmd == nil {
		t.Fatal("root command missing layout-template help command")
	}
	if cmd.Use != "layout-template" {
		t.Errorf("layout-template command Use = %q, want layout-template", cmd.Use)
	}
}

func TestLayoutTemplateCommandOutputCoversFieldsAndSafety(t *testing.T) {
	var output bytes.Buffer
	layoutTemplateCmd.SetOut(&output)
	t.Cleanup(func() {
		layoutTemplateCmd.SetOut(nil)
	})

	layoutTemplateCmd.Run(layoutTemplateCmd, nil)

	got := output.String()
	for _, want := range []string{
		"{author}",
		"{book_title}",
		"{series-count}",
		"{narrators}",
		"{publisher-name}",
		"{series|Standalone}",
		"{Vol series_number:02 - }",
		"${field}",
		"Composite optional segment",
		"Absolute templates",
		"https://github.com/jeeftor/audiobook-organizer/blob/master/docs/LAYOUTS.md",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("layout-template output missing %q", want)
		}
	}
	if strings.Contains(got, "\n  docs/LAYOUTS.md\n") {
		t.Error("layout-template output still uses only the local docs path")
	}
}

func TestLayoutTemplateCommandSuppressesStartupBanner(t *testing.T) {
	if shouldPrintStartupBanner([]string{"layout-template"}) {
		t.Fatal("layout-template reference command should suppress startup banner")
	}
}

func TestABSOrganizeCommandIncludesLayoutTemplateFlag(t *testing.T) {
	flag := absOrganizeCmd.Flags().Lookup("layout-template")
	if flag == nil {
		t.Fatal("abs organize command missing layout-template flag")
	}
	if flag.DefValue != "" {
		t.Errorf("layout-template default = %q, want empty", flag.DefValue)
	}
}
