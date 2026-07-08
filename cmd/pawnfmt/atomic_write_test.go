package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteCreatesANewFile(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "new.pwn")
	if err := atomicWrite(path, []byte("new x;\n")); err != nil {
		t.Fatalf("atomicWrite: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	if string(got) != "new x;\n" {
		t.Fatalf("written content = %q, want %q", got, "new x;\n")
	}
}

func TestAtomicWriteOverwritesAnExistingFilePreservingItsMode(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "existing.pwn")
	if err := os.WriteFile(path, []byte("old\n"), 0o640); err != nil {
		t.Fatalf("seed existing file: %v", err)
	}

	if err := atomicWrite(path, []byte("new\n")); err != nil {
		t.Fatalf("atomicWrite: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}

	if string(got) != "new\n" {
		t.Fatalf("written content = %q, want %q", got, "new\n")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}

	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode after overwrite = %v, want the original 0640 preserved", info.Mode().Perm())
	}
}

func TestAtomicWriteLeavesNoTempFileBehindOnSuccess(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "clean.pwn")
	if err := atomicWrite(path, []byte("x\n")); err != nil {
		t.Fatalf("atomicWrite: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir: %v", err)
	}

	if len(entries) != 1 || entries[0].Name() != "clean.pwn" {
		t.Fatalf("directory contents after atomicWrite = %v, want only clean.pwn (no leftover .pawnfmt-*.tmp)", entries)
	}
}

func TestAtomicWriteFailsCleanlyWhenTheDirectoryDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-dir", "file.pwn")
	if err := atomicWrite(path, []byte("x\n")); err == nil {
		t.Fatal("atomicWrite should fail when its target directory does not exist")
	}
}
