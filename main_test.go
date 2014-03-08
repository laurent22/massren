package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setup(t *testing.T) {
	minLogLevel_ = 10

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	homeDir_ = filepath.Join(pwd, "homedirtest")
	err = os.MkdirAll(homeDir_, 0700)
	if err != nil {
		t.Fatal(err)
	}

	deleteTempFiles()
	profileOpen()
	clearHistory()
}

func teardown(t *testing.T) {
	profileDelete()
}

func touch(filePath string) {
	ioutil.WriteFile(filePath, []byte("testing"), 0700)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func createRandomTempFiles() []string {
	var output []string
	for i := 0; i < 10; i++ {
		filePath := filepath.Join(tempFolder(), fmt.Sprintf("testfile%d", i))
		ioutil.WriteFile(filePath, []byte("testing"), 0700)
		output = append(output, filePath)
	}
	return output
}

func Test_stringHash(t *testing.T) {
	if len(stringHash("aaaa")) != 32 {
		t.Error("hash should be 32 characters long")
	}

	if stringHash("abcd") == stringHash("efgh") || stringHash("") == stringHash("ijkl") {
		t.Error("hashes should be different")
	}
}

func Test_watchFile(t *testing.T) {
	setup(t)
	defer teardown(t)

	filePath := tempFolder() + "watchtest"
	ioutil.WriteFile(filePath, []byte("testing"), 0700)
	doneChan := make(chan bool)

	go func(doneChan chan bool) {
		defer func() {
			doneChan <- true
		}()
		err := watchFile(filePath)
		if err != nil {
			t.Error(err)
		}
	}(doneChan)

	time.Sleep(300 * time.Millisecond)
	ioutil.WriteFile(filePath, []byte("testing change"), 0700)

	<-doneChan
}

func fileListsAreEqual(files1 []string, files2 []string) error {
	if len(files1) != len(files2) {
		return errors.New("file count is different")
	}

	for _, f1 := range files1 {
		found := false
		for _, f2 := range files2 {
			if filepath.Base(f1) == filepath.Base(f2) {
				found = true
			}
		}
		if !found {
			return errors.New("file names do not match")
		}
	}

	return nil
}

func Test_filePathsFromArgs(t *testing.T) {
	setup(t)
	defer teardown(t)

	tempFiles := createRandomTempFiles()
	args := []string{
		filepath.Join(tempFolder(), "*"),
	}

	filePaths, err := filePathsFromArgs(args)
	if err != nil {
		t.Fatal(err)
	}

	err = fileListsAreEqual(filePaths, tempFiles)
	if err != nil {
		t.Error(err)
	}

	// If no argument is provided, the function should default to "*"
	// in the current dir.

	currentDir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	err = os.Chdir(tempFolder())
	if err != nil {
		panic(err)
	}

	args = []string{}
	filePaths, err = filePathsFromArgs(args)
	if err != nil {
		t.Fatal(err)
	}

	err = fileListsAreEqual(filePaths, tempFiles)
	if err != nil {
		t.Error(err)
	}

	os.Chdir(currentDir)
}

func stringListsEqual(s1 []string, s2 []string) bool {
	for i, s := range s1 {
		if s != s2[i] {
			return false
		}
	}
	return true
}

func Test_filePathsFromString(t *testing.T) {
	newline_ = "\n"

	var data []string
	var expected [][]string

	data = append(data, "// comment\n\nfile1\nfile2\n//comment\n\n\n")
	expected = append(expected, []string{"file1", "file2"})

	data = append(data, "\n// comment\n\n")
	expected = append(expected, []string{})

	data = append(data, "")
	expected = append(expected, []string{})

	data = append(data, "// comment\n\n  file1 \n\tfile2\n\nanother file\t\n//comment\n\n\n")
	expected = append(expected, []string{"  file1 ", "\tfile2", "another file\t"})

	for i, d := range data {
		e := expected[i]
		r := filePathsFromString(d)
		if !stringListsEqual(e, r) {
			t.Error("Expected", e, "got", r)
		}
	}
}

func Test_filePathsFromListFile(t *testing.T) {
	setup(t)
	defer teardown(t)

	ioutil.WriteFile(filepath.Join(tempFolder(), "list.txt"), []byte("one"+newline()+"two"), PROFILE_PERM)
	filePaths, err := filePathsFromListFile(filepath.Join(tempFolder(), "list.txt"))
	if err != nil {
		t.Errorf("Expected no error, got %s", err)
	}

	if len(filePaths) != 2 {
		t.Errorf("Expected 2 paths, got %d", len(filePaths))
	} else {
		if filePaths[0] != "one" || filePaths[1] != "two" {
			t.Error("Incorrect data")
		}
	}

	os.Remove(filepath.Join(tempFolder(), "list.txt"))
	_, err = filePathsFromListFile(filepath.Join(tempFolder(), "list.txt"))
	if err == nil {
		t.Error("Expected an error, got nil")
	}
}

func Test_stripBom(t *testing.T) {
	data := [][][]byte{
		{{239, 187, 191}, {}},
		{{239, 187, 191, 239, 187, 191}, {239, 187, 191}},
		{{239, 187, 191, 65, 66}, {65, 66}},
		{{239, 191, 65, 66}, {239, 191, 65, 66}},
		{{}, {}},
		{{65, 239, 187, 191}, {65, 239, 187, 191}},
	}

	for _, d := range data {
		if stripBom(string(d[0])) != string(d[1]) {
			t.Errorf("Expected %x, got %x", d[0], d[1])
		}
	}
}

func Test_deleteTempFiles(t *testing.T) {
	setup(t)
	defer teardown(t)

	ioutil.WriteFile(filepath.Join(tempFolder(), "one"), []byte("test1"), PROFILE_PERM)
	ioutil.WriteFile(filepath.Join(tempFolder(), "two"), []byte("test2"), PROFILE_PERM)

	deleteTempFiles()

	tempFiles, _ := filepath.Glob(filepath.Join(tempFolder(), "*"))
	if len(tempFiles) > 0 {
		t.Fail()
	}
}

func Test_renameFiles(t *testing.T) {
	setup(t)
	defer teardown(t)

	touch(filepath.Join(tempFolder(), "one"))
	touch(filepath.Join(tempFolder(), "two"))
	touch(filepath.Join(tempFolder(), "three"))

	hasChanges, _, _ := renameFiles([]string{filepath.Join(tempFolder(), "one"), filepath.Join(tempFolder(), "two")}, []string{"one123", "two456"}, false)

	if !hasChanges {
		t.Error("Expected changes.")
	}

	if !fileExists(filepath.Join(tempFolder(), "one123")) {
		t.Error("File not found")
	}

	if !fileExists(filepath.Join(tempFolder(), "two456")) {
		t.Error("File not found")
	}

	if !fileExists(filepath.Join(tempFolder(), "three")) {
		t.Error("File not found")
	}

	renameFiles([]string{filepath.Join(tempFolder(), "three")}, []string{"nochange"}, true)

	if !fileExists(filepath.Join(tempFolder(), "three")) {
		t.Error("File was renamed in dry-run mode")
	}

	hasChanges, _, _ = renameFiles([]string{filepath.Join(tempFolder(), "three")}, []string{"three"}, false)
	if hasChanges {
		t.Error("No file should have been renamed")
	}
}

func Test_newline(t *testing.T) {
	newline_ = ""
	nl := newline()
	if len(nl) < 1 || len(nl) > 2 {
		t.Fail()
	}
}

func Test_guessEditorCommand(t *testing.T) {
	editor, err := guessEditorCommand()
	if err != nil || len(editor) <= 0 {
		t.Fail()
	}
}

func Test_duplicatePaths(t *testing.T) {
	if len(duplicatePaths([]string{"one", "two", "one", "three"})) != 1 {
		t.Error("Wrong duplicate count")
	}

	if len(duplicatePaths([]string{"one", "two", "three"})) != 0 {
		t.Error("Wrong duplicate count")
	}

	if len(duplicatePaths([]string{})) != 0 {
		t.Error("Wrong duplicate count")
	}
}
