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

var createdTempFiles []string

func setup(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	homeDir_ = pwd + "/homedirtest"
	err = os.MkdirAll(homeDir_, 0700)
	if err != nil {
		t.Fatal(err)
	}
}

func teardown(t *testing.T) {
	for _, filePath := range createdTempFiles {
		os.Remove(filePath)
	}	
}

func createRandomTempFiles() []string {
	var output []string
	for i := 0; i < 10; i++ {
		filePath := fmt.Sprintf("%s/testfile%d", tempFolder(), i)
		ioutil.WriteFile(filePath, []byte("testing"), 0700)
		output = append(output, filePath)
		createdTempFiles = append(createdTempFiles, filePath)
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

func Test_configFolder(t *testing.T) {
	setup(t)
	defer teardown(t)
	
	configFolder := configFolder()
	stat, err := os.Stat(configFolder)
	if err != nil {
		t.Error(err)
	}
	if !stat.IsDir() {
		t.Error("config folder is not a directory: %s", configFolder)
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
	
	time.Sleep(500 * time.Millisecond)
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