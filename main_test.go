package main

import (
	"io/ioutil"
	"os"
	"testing"
	"time"
)

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
	
}

func TestHash(t *testing.T) {
	if len(stringHash("aaaa")) != 32 {
		t.Error("hash should be 32 characters long")
	}
	
	if stringHash("abcd") == stringHash("efgh") || stringHash("") == stringHash("ijkl") {
		t.Error("hashes should be different")
	}
}

func TestConfigFolder(t *testing.T) {
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

func TestWatchFile(t *testing.T) {
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