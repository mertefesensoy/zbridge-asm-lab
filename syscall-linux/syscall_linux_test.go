//go:build linux && amd64

package syscalllinux

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestGetpidMatchesOS(t *testing.T) {
	pid, errno := Getpid()
	if errno != 0 {
		t.Fatalf("Getpid errno = %d, want 0", errno)
	}
	if want := uintptr(os.Getpid()); pid != want {
		t.Fatalf("Getpid = %d, want %d", pid, want)
	}
}

func TestWriteToPipe(t *testing.T) {
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	msg := []byte("syscall-linux\n")
	n, errno := Write(w.Fd(), msg)
	if errno != 0 {
		w.Close()
		t.Fatalf("Write errno = %d, want 0", errno)
	}
	if n != uintptr(len(msg)) {
		w.Close()
		t.Fatalf("Write returned n = %d, want %d", n, len(msg))
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, msg) {
		t.Fatalf("pipe contents = %q, want %q", got, msg)
	}
}

func TestWriteBadFDReturnsErrno(t *testing.T) {
	n, errno := Write(^uintptr(0), []byte("x"))
	if errno == 0 {
		t.Fatalf("Write bad fd errno = 0, want non-zero")
	}
	if n != ^uintptr(0) {
		t.Fatalf("Write bad fd n = %d, want ^uintptr(0)", n)
	}
}
