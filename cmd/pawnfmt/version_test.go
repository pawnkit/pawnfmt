package main

import "testing"

func TestVersionStringIsNeverEmpty(t *testing.T) {
	if got := versionString(); got == "" {
		t.Fatal("versionString() returned an empty string")
	}
}

func TestRunVersionFlagPrintsVersionAndExitsOK(t *testing.T) {
	code, stdout, stderr := runCLI([]string{"--version"}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}
	if stdout == "" {
		t.Fatal("-version printed nothing to stdout")
	}
}

func TestRunVersionFlagTakesPrecedenceOverOtherFlags(t *testing.T) {
	code, stdout, stderr := runCLI([]string{"--version", "--write", "--check"}, "")
	if code != exitOK {
		t.Fatalf("exit code = %d, want %d (stderr: %s)", code, exitOK, stderr)
	}
	if stdout == "" {
		t.Fatal("-version printed nothing to stdout")
	}
}
