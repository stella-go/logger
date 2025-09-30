// Copyright 2010-2025 the original author or authors.

package logger

import "testing"

func TestSyslogWriter(t *testing.T) {
	w, err := NewSyslogWriter("localhost:514")
	if err != nil {
		t.Fatal(err)
	}
	_, err = w.Write([]byte("test syslog message"))
	if err != nil {
		t.Fatal(err)
	}
}
