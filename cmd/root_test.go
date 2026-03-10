package cmd

import "testing"

func TestRootCommands(t *testing.T) {
	if _, _, err := rootCmd.Find([]string{"create"}); err != nil {
		t.Fatalf("expected create command to be registered: %v", err)
	}

	if _, _, err := rootCmd.Find([]string{"update"}); err != nil {
		t.Fatalf("expected update command to be registered: %v", err)
	}

	if _, _, err := rootCmd.Find([]string{"edit"}); err == nil {
		t.Fatal("expected edit command to be removed")
	}

	if _, _, err := rootCmd.Find([]string{"search"}); err != nil {
		t.Fatalf("expected search command to be registered: %v", err)
	}

	if _, _, err := rootCmd.Find([]string{"nt"}); err == nil {
		t.Fatal("expected nt command to be removed")
	}
}
