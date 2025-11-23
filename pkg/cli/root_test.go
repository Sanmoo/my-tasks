package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestNewRootCmd(t *testing.T) {
	rootCmd := NewRootCmd()

	if rootCmd.Use != "tasks" {
		t.Errorf("Expected command use 'tasks', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short != "A simple and powerful CLI task manager" {
		t.Errorf("Expected short description '%s', got '%s'", "A simple and powerful CLI task manager", rootCmd.Short)
	}

	if rootCmd.Long != "A task manager CLI tailored to my own needs" {
		t.Errorf("Expected long description '%s', got '%s'", "A task manager CLI tailored to my own needs", rootCmd.Long)
	}

	// Check that subcommands are added
	subcommands := rootCmd.Commands()

	// Debug: print what subcommands we found
	for i, cmd := range subcommands {
		t.Logf("Subcommand %d: Use='%s', Name='%s'", i, cmd.Use, cmd.Name())
	}

	if len(subcommands) != 2 {
		t.Errorf("Expected 2 subcommands, got %d", len(subcommands))
	}

	// Verify specific subcommands exist
	hasList := false
	hasRemind := false
	for _, cmd := range subcommands {
		if cmd.Name() == "list" {
			hasList = true
		}
		if cmd.Name() == "remind" {
			hasRemind = true
		}
	}

	if !hasList {
		t.Error("Expected 'list' subcommand to be present")
	}
	if !hasRemind {
		t.Error("Expected 'remind' subcommand to be present")
	}
}

func TestRootCmd_PersistentPreRunE(t *testing.T) {
	// Create a test command to verify app initialization
	testCmd := &cobra.Command{
		Use: "test",
	}

	// Add the persistent pre-run from root command
	rootCmd := NewRootCmd()
	testCmd.PersistentPreRunE = rootCmd.PersistentPreRunE

	// Execute the pre-run to verify app initialization
	err := testCmd.PersistentPreRunE(testCmd, []string{})

	// This should succeed and initialize the App
	if err != nil {
		t.Errorf("Expected no error from PersistentPreRunE, got: %v", err)
	}

	if App == nil {
		t.Error("Expected App to be initialized after PersistentPreRunE")
	}
}

func TestExecute_FunctionSignature(t *testing.T) {
	// Test that Execute function exists and has the correct signature
	// This is a basic smoke test to ensure the function can be called
	// We can't easily test the actual execution without complex mocking
	// but we can verify the function is defined and accessible

	// The function should be callable without panicking
	// We'll just verify it exists by checking the package exports it
	// Execute is a function variable, so we can't test for nil
	// This test mainly ensures the package compiles correctly
}
