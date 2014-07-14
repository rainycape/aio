package aio

import (
	"bytes"
	"os"
	"testing"
)

func TestAIO(t *testing.T) {
	const (
		foobar = "foobar"
	)
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	var b [2048]byte
	err = s.Add(r, In, nil, func(_ *Event) {
		n, err := r.Read(b[:])
		if err != nil {
			t.Fatal(err)
		}
		buf.Write(b[:n])
	})
	if err != nil {
		t.Fatal(err)
	}
	w.Write([]byte(foobar))
	s.Wait(-1)
	if buf.String() != foobar {
		t.Errorf("expecting string %s, got %s instead", foobar, buf.String())
	}
	if err := s.Delete(r); err != nil {
		t.Fatal(err)
	}
	// This write should not be notified
	w.Write([]byte(foobar))
	s.Wait(0)
	if buf.String() != foobar {
		t.Errorf("expecting string %s, got %s instead", foobar, buf.String())
	}
}
