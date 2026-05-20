package cmd

import "testing"

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

func TestABSOrganizeCommandIncludesLayoutTemplateFlag(t *testing.T) {
	flag := absOrganizeCmd.Flags().Lookup("layout-template")
	if flag == nil {
		t.Fatal("abs organize command missing layout-template flag")
	}
	if flag.DefValue != "" {
		t.Errorf("layout-template default = %q, want empty", flag.DefValue)
	}
}
