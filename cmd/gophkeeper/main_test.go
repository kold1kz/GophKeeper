package main

import (
	"flag"
	"os"
	"testing"
)

func resetFlags() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
}

func TestNa(t *testing.T) {
	if got := na(""); got != "N/A" {
		t.Fatalf("expected N/A, got %q", got)
	}
	if got := na("abc"); got != "abc" {
		t.Fatalf("expected abc, got %q", got)
	}
}
