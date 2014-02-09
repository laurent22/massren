package main

import (
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
}