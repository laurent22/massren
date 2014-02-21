package main

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_userHomeDir(t *testing.T) {
	homeDir := userHomeDir()
	if len(homeDir) == 0 {
		t.Fail()
	}
}

func Test_profileOpenClose(t *testing.T) {
	if profileDb_ != nil {
		t.Error("profileDb_ should be nil")
	}

	pwd, _ := os.Getwd()
	homeDir_ = filepath.Join(pwd, "homedirtest")

	err := profileOpen()
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if profileDb_ == nil {
		t.Error("profileDb_ should be set")
	}

	profileClose()
	if profileDb_ != nil {
		t.Error("profileDb_ should be nil")
	}

	profileDelete()
}

func Test_profileFolder(t *testing.T) {
	setup(t)
	defer teardown(t)

	profileFolder := profileFolder()
	stat, err := os.Stat(profileFolder)
	if err != nil {
		t.Error(err)
	}

	if !stat.IsDir() {
		t.Error("config folder is not a directory: %s", profileFolder)
	}
}

func Test_handleConfigCommand_noArgs(t *testing.T) {
	setup(t)
	defer teardown(t)

	var opts CommandLineOptions
	var err error

	err = handleConfigCommand(&opts, []string{})
	if err != nil {
		t.Error("Expected no error")
	}

	err = handleConfigCommand(&opts, []string{"testing", "123"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if config_.String("testing") != "123" {
		t.Error("Value not set correctly")
	}

	handleConfigCommand(&opts, []string{"testing", "abcd"})
	if config_.String("testing") != "abcd" {
		t.Error("Value has not been changed")
	}

	err = handleConfigCommand(&opts, []string{"testing"})
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if config_.HasKey("testing") {
		t.Error("Key has not been deleted")
	}
}
