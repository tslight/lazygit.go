package common

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfDir(t *testing.T) {
	var want string
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	switch runtime.GOOS {
	case "linux", "freebsd", "openbsd":
		want = filepath.Join(home, ".config/lazygit")
	case "windows":
		want = filepath.Join(home, "/AppData/Roaming/lazygit")
	case "darwin":
		want = filepath.Join(home, "/Library/Application Support/lazygit")
	}
	got := ConfDir()
	if got != want {
		t.Fatalf(`ConfDir() returned %v, instead of %v`, got, want)
	}
}

func TestAbsHomeDirTilde(t *testing.T) {
	want, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	tilde := "~"
	got := AbsHomeDir(tilde)
	if got != want {
		t.Fatalf(`AbsHomeDir(%s) returned %v, instead of %v`, tilde, got, want)
	}
}

func TestAbsHomeDirHOME(t *testing.T) {
	want, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	h := "$HOME"
	got := AbsHomeDir(h)
	if got != want {
		t.Fatalf(`AbsHomeDir(%s) returned %v, instead of %v`, h, got, want)
	}
}

func TestAbsHomeDirUSERPROFILE(t *testing.T) {
	want, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	up := "%USERPROFILE%"
	got := AbsHomeDir(up)
	if got != want {
		t.Fatalf(`AbsHomeDir(%s) returned %v, instead of %v`, up, got, want)
	}
}
