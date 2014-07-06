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

	var fileActions []*FileAction
	var fileAction *FileAction

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "one")
	fileAction.newPath = "123"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "two")
	fileAction.newPath = "456"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{
		filepath.Join(tempFolder(), "123"), filepath.Join(tempFolder(), "456"),
	})

	if err != nil {
		t.Errorf("Expected no error, got: %s", err)
	}

	if !fileExists(filepath.Join(tempFolder(), "one")) || !fileExists(filepath.Join(tempFolder(), "two")) {
		t.Error("Undo operation did not restore filenames")
	}

	historyItems, _ := allHistoryItems()
	if len(historyItems) > 0 {
		t.Error("Undo operation did not delete restored history.")
	}

	fileActions = []*FileAction{}

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "one")
	fileAction.newPath = "123"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "two")
	fileAction.newPath = "456"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

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

	var fileActions []*FileAction
	var fileAction *FileAction

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "one")
	fileAction.newPath = "123"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = filepath.Join(tempFolder(), "two")
	fileAction.newPath = "456"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

	os.Remove(filepath.Join(tempFolder(), "123"))

	var opts CommandLineOptions
	err = handleUndoCommand(&opts, []string{
		filepath.Join(tempFolder(), "123"),
	})

	if err == nil {
		t.Fail()
	}
}

func Test_handleUndoCommand_withIntermediateRename(t *testing.T) {
	setup(t)
	defer teardown(t)

	p0 := filepath.Join(tempFolder(), "0")
	p1 := filepath.Join(tempFolder(), "1")

	filePutContent(p0, "0")
	filePutContent(p1, "1")

	fileActions := []*FileAction{}

	fileAction := NewFileAction()
	fileAction.oldPath = p0
	fileAction.newPath = "1"
	fileActions = append(fileActions, fileAction)

	fileAction = NewFileAction()
	fileAction.oldPath = p1
	fileAction.newPath = "0"
	fileActions = append(fileActions, fileAction)

	processFileActions(fileActions, false)

	var opts CommandLineOptions
	err := handleUndoCommand(&opts, []string{
		p0, p1,
	})

	if err != nil {
		t.Fail()
	}

	if fileGetContent(p0) != "0" {
		t.Error("File 0 was not restored")
	}

	if fileGetContent(p1) != "1" {
		t.Error("File 1 was not restored")
	}
}
