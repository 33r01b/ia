package app

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFilterHostMountTargetsKeepsExistingRelativeTargets(t *testing.T) {
	workdir := t.TempDir()
	withWorkingDir(t, workdir)

	if err := os.MkdirAll(filepath.Join("billing", ".cache"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("billing", ".env"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var targets MountTargets
	targets.Add(".env")
	targets.Add(".cache")

	got := filterHostMountTargets("billing", targets).Items()
	want := []string{".env", ".cache"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected targets: got %#v want %#v", got, want)
	}
}

func TestFilterHostMountTargetsSkipsMissingRelativeTargets(t *testing.T) {
	workdir := t.TempDir()
	withWorkingDir(t, workdir)

	if err := os.MkdirAll("billing", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join("billing", ".env"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var targets MountTargets
	targets.Add(".env")
	targets.Add(".missing")
	targets.Add("tmp/runtime")

	got := filterHostMountTargets("billing", targets).Items()
	want := []string{".env"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected targets: got %#v want %#v", got, want)
	}
}

func TestFilterHostMountTargetsKeepsAbsoluteTargets(t *testing.T) {
	workdir := t.TempDir()
	withWorkingDir(t, workdir)

	if err := os.MkdirAll("billing", 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	var targets MountTargets
	targets.Add("/app/go/billing/.env")
	targets.Add("/app/go/billing/.cache")

	got := filterHostMountTargets("billing", targets).Items()
	want := []string{"/app/go/billing/.env", "/app/go/billing/.cache"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected targets: got %#v want %#v", got, want)
	}
}

func withWorkingDir(t *testing.T, dir string) {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to temp dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore working dir: %v", err)
		}
	})
}
