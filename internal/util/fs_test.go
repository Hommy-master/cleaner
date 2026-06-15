package util

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsAbsolutePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "unix absolute", path: "/tmp/foo", want: runtime.GOOS != "windows"},
		{name: "windows drive absolute", path: `D:\apps\foo`, want: runtime.GOOS == "windows"},
		{name: "relative segment", path: "foo/bar", want: false},
		{name: "dot relative", path: "./foo", want: false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := IsAbsolutePath(tt.path); got != tt.want {
				t.Fatalf("IsAbsolutePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidateDirPath(t *testing.T) {
	t.Parallel()

	if err := ValidateDirPath(""); err == nil {
		t.Fatal("expected error for empty path")
	}
	if err := ValidateDirPath("   "); err == nil {
		t.Fatal("expected error for whitespace path")
	}

	if runtime.GOOS == "windows" {
		if err := ValidateDirPath(`D:`); err == nil {
			t.Fatal("expected error for short windows path")
		}
		if err := ValidateDirPath(`D:\a`); err != nil {
			t.Fatalf("valid windows path rejected: %v", err)
		}
	} else {
		if err := ValidateDirPath("/"); err == nil {
			t.Fatal("expected error for single-char linux path")
		}
		if err := ValidateDirPath("/tmp"); err != nil {
			t.Fatalf("valid linux path rejected: %v", err)
		}
	}
}

func TestClassifyPath(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	kind, err := ClassifyPath(file)
	if err != nil {
		t.Fatalf("classify file: %v", err)
	}
	if kind != PathFile {
		t.Fatalf("file kind = %v, want PathFile", kind)
	}

	kind, err = ClassifyPath(dir)
	if err != nil {
		t.Fatalf("classify dir: %v", err)
	}
	if kind != PathDirectory {
		t.Fatalf("dir kind = %v, want PathDirectory", kind)
	}

	kind, err = ClassifyPath(filepath.Join(dir, "missing"))
	if err != nil {
		t.Fatalf("classify missing: %v", err)
	}
	if kind != PathMissing {
		t.Fatalf("missing kind = %v, want PathMissing", kind)
	}
}

func TestNormalizeRelativePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{in: "a\\b", want: "a/b"},
		{in: "./foo", want: "foo"},
		{in: ".", want: ""},
	}

	for _, tt := range tests {
		if got := NormalizeRelativePath(tt.in); got != tt.want {
			t.Fatalf("NormalizeRelativePath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestIsIgnored(t *testing.T) {
	t.Parallel()

	ignore := []string{"filename", "dir/test.exe", "keepdir"}

	tests := []struct {
		rel  string
		want bool
	}{
		{rel: "filename", want: true},
		{rel: "dir/test.exe", want: true},
		{rel: "dir/other.exe", want: false},
		{rel: "keepdir/file.txt", want: true},
		{rel: "keepdir/nested/x", want: true},
		{rel: "remove.me", want: false},
	}

	for _, tt := range tests {
		if got := IsIgnored(tt.rel, ignore); got != tt.want {
			t.Fatalf("IsIgnored(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}
