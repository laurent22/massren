package main

import (
	"os"
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

	touch(tempFolder() + "/one")
	touch(tempFolder() + "/two")
	touch(tempFolder() + "/three")
	
	renameFiles(
		[]string{tempFolder() + "/one", tempFolder() + "/two"}, 
		[]string{"123", "456"},
		false,
	)
	
	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{
		tempFolder() + "/123", tempFolder() + "/456", 		
	})
	
	if err != nil {
		t.Errorf("Expected not error, got %s", err)
	}
	
	if !fileExists(tempFolder() + "/one") || !fileExists(tempFolder() + "/two") {
		t.Error("Undo operation did not restore filenames")
	}
	
	historyItems, _ := allHistoryItems()
	if len(historyItems) > 0 {
		t.Error("Undo operation did not delete restored history.")
	}
	
	renameFiles(
		[]string{tempFolder() + "/one", tempFolder() + "/two"}, 
		[]string{"123", "456"},
		false,
	)

	opts = CommandLineOptions{
		DryRun: true,
	}
	err = handleUndoCommand(&opts, []string{
		tempFolder() + "/123", tempFolder() + "/456", 		
	})

	if !fileExists(tempFolder() + "/123") || !fileExists(tempFolder() + "/456") {
		t.Error("Undo operation in dry run mode restored filenames.")
	}
}

func Test_handleUndoCommand_fileHasBeenDeleted(t *testing.T) {
	setup(t)
	defer teardown(t)

	var err error

	touch(tempFolder() + "/one")
	touch(tempFolder() + "/two")
	touch(tempFolder() + "/three")
	
	renameFiles(
		[]string{tempFolder() + "/one", tempFolder() + "/two"}, 
		[]string{"123", "456"},
		false,
	)

	os.Remove(tempFolder() + "/123")

	var opts CommandLineOptions
	err = handleUndoCommand(&opts, []string{
		tempFolder() + "/123", 		
	})
	
	if err == nil {
		t.Fail()
	}
}