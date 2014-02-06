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
	minLogLevel_ = 1
	
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	homeDir_ = pwd + "/homedirtest"
	err = os.MkdirAll(homeDir_, 0700)
	if err != nil {
		t.Fatal(err)
	}
		
	deleteTempFiles()
	
	profileOpen()
}

func teardown(t *testing.T) {
	profileClose()
	deleteTempFiles()
}

func createRandomTempFiles() []string {
	var output []string
	for i := 0; i < 10; i++ {
		filePath := fmt.Sprintf("%s/testfile%d", tempFolder(), i)
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
		tempFolder() + "/*",
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

func Test_stripBom(t *testing.T) {
	data := [][][]byte{
		[][]byte{ []byte{239,187,191}, []byte{} },
		[][]byte{ []byte{239,187,191,239,187,191}, []byte{239,187,191} },
		[][]byte{ []byte{239,187,191,65,66}, []byte{65,66} },
		[][]byte{ []byte{239,191,65,66}, []byte{239,191,65,66} },
		[][]byte{ []byte{}, []byte{} },
		[][]byte{ []byte{65,239,187,191}, []byte{65,239,187,191} },
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
	
	ioutil.WriteFile(tempFolder() + "/one", []byte("test1"), PROFILE_PERM)
	ioutil.WriteFile(tempFolder() + "/two", []byte("test2"), PROFILE_PERM)
	
	deleteTempFiles()
	
	tempFiles, _ := filepath.Glob(tempFolder() + "/*")
	if len(tempFiles) > 0 {
		t.Fail()
	}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

func Test_renameFiles(t *testing.T) {
	setup(t)
	defer teardown(t)
	
	ioutil.WriteFile(tempFolder() + "/one", []byte("1"), PROFILE_PERM)
	ioutil.WriteFile(tempFolder() + "/two", []byte("2"), PROFILE_PERM)
	ioutil.WriteFile(tempFolder() + "/three", []byte("3"), PROFILE_PERM)
	
	hasChanges, _, _ := renameFiles([]string{tempFolder() + "/one", tempFolder() + "/two"}, []string{"one123", "two456"}, false)
	
	if !hasChanges {
		t.Error("Expected changes.")
	}
	
	if !fileExists(tempFolder() + "/one123") {
		t.Error("File not found")
	}

	if !fileExists(tempFolder() + "/two456") {
		t.Error("File not found")
	}
	
	if !fileExists(tempFolder() + "/three") {
		t.Error("File not found")
	}
}