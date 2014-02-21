package main

import (
	"testing"
)

func Test_handleVersionCommand(t *testing.T) {
	var opts CommandLineOptions
	err := handleVersionCommand(&opts, []string{})
	if err != nil {
		t.Fail()
	}
}