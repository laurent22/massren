package main

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_handleUndoCommand_noArgs(t *testing.T) {
	setup(t)
	defer teardown(t)

	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{})
	if err != nil {
		t.Fail()
	}
}

func Test_handleUndoCommand_notInHistory(t *testing.T) {
	setup(t)
	defer teardown(t)

	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{"one", "two"})
	if err != nil {
		t.Fail()
	}
}

func Test_handleUndoCommand(t *testing.T) {
	setup(t)
	defer teardown(t)

	touch(filepath.Join(tempFolder(), "one"))
	touch(filepath.Join(tempFolder(), "two"))
	touch(filepath.Join(tempFolder(), "three"))

	renameFiles(
		[]string{filepath.Join(tempFolder(), "one"), filepath.Join(tempFolder(), "two")},
		[]string{"123", "456"},
		false,
	)

	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{
		filepath.Join(tempFolder(), "123"), filepath.Join(tempFolder(), "456"),
	})

	if err != nil {
		t.Errorf("Expected not error, got %s", err)
	}

	if !fileExists(filepath.Join(tempFolder(), "one")) || !fileExists(filepath.Join(tempFolder(), "two")) {
		t.Error("Undo operation did not restore filenames")
	}

	historyItems, _ := allHistoryItems()
	if len(historyItems) > 0 {
		t.Error("Undo operation did not delete restored history.")
	}

	renameFiles(
		[]string{filepath.Join(tempFolder(), "one"), filepath.Join(tempFolder(), "two")},
		[]string{"123", "456"},
		false,
	)

	opts = CommandLineOptions{
		DryRun: true,
	}
	err = handleUndoCommand(&opts, []string{
		filepath.Join(tempFolder(), "123"), filepath.Join(tempFolder(), "456"),
	})

	if !fileExists(filepath.Join(tempFolder(), "123")) || !fileExists(filepath.Join(tempFolder(), "456")) {
		t.Error("Undo operation in dry run mode restored filenames.")
	}
}

func Test_handleUndoCommand_fileHasBeenDeleted(t *testing.T) {
	setup(t)
	defer teardown(t)

	var err error

	touch(filepath.Join(tempFolder(), "one"))
	touch(filepath.Join(tempFolder(), "two"))
	touch(filepath.Join(tempFolder(), "three"))

	renameFiles(
		[]string{filepath.Join(tempFolder(), "one"), filepath.Join(tempFolder(), "two")},
		[]string{"123", "456"},
		false,
	)

	os.Remove(filepath.Join(tempFolder(), "123"))

	var opts CommandLineOptions
	err = handleUndoCommand(&opts, []string{
		filepath.Join(tempFolder(), "123"),
	})

	if err == nil {
		t.Fail()
	}
}
