package main

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"github.com/jessevdk/go-flags"	
	"sort"
	"time"
)

const PROGRAM_NAME = "massren"

var homeDir_ string
var configFolder_ string

// TODO: catch SIGTERM

func stringHash(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))	
}

func userHomeDir() string {
	u, err := user.Current()
	if err != nil {
		panic(err)
	}
	return u.HomeDir	
}

func configFolder() string {
	if configFolder_ != "" {
		return configFolder_
	}
	
	if homeDir_ == "" {
		u, err := user.Current()
		if err != nil {
			panic(err)
		}
		homeDir_ = u.HomeDir
	}
	
	output := homeDir_ + "/.config/massren"
	
	err := os.MkdirAll(output, 0700)
	if err != nil {
		panic(err)
	}
	
	configFolder_ = output
	return configFolder_
}

func tempFolder() string {
	output := configFolder() + "/temp"
	err := os.MkdirAll(output, 0700)
	if err != nil {
		panic(err)
	}
	return output
}

func criticalError(err error) {
	fmt.Println(err)
	fmt.Printf("Run '%s --help' for usage\n", PROGRAM_NAME) 
	os.Exit(1)
}

func watchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}
		
		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			return nil
		}
		
		time.Sleep(1 * time.Second)
	}
	
	panic("unreachable")
}

func editFile(filePath string) error {
	cmd := exec.Command("sub", filePath)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return errors.New(fmt.Sprintf("%s: %s", err, stderr.String()))
	}
	return nil
}

func filePathsFromArgs(args []string) ([]string, error) {
	var output []string
	var err error
	
	if len(args) == 0 {
		output, err = filepath.Glob("*")
		if err != nil {
			return []string{}, err
		}
	} else {
		for _, arg := range args {
			matches, err := filepath.Glob(arg)
			if err != nil {
				return []string{}, err
			}
			for _, match := range matches {
				output = append(output, match)
			}
		}
	}
	
	sort.Strings(output)
	
	return output, nil
}

func main() {
	var opts struct {
		DryRun bool `short:"n" long:"dry-run" description:"Don't rename anything but show the operation that would have been performed."`
	}

	args, err := flags.Parse(&opts)
	if err != nil {
		if err.(*flags.Error).Type == flags.ErrHelp {
			return
		}
		criticalError(err)
	}

	filePaths, err := filePathsFromArgs(args)
	if err != nil {
		criticalError(err)
	}
	
	listFileContent := ""
	md5FileContent := ""
	baseFilename := ""
	for _, filePath := range filePaths {
		listFileContent += filePath + "\n"
		md5FileContent += stringHash(filePath) + "\n"
		baseFilename += md5FileContent + "_"
	}
	
	baseFilename = stringHash(baseFilename)
	listFilePath := configFolder() + "/" + baseFilename + ".files.txt"
	md5FilePath := configFolder() + "/" + baseFilename + ".md5.txt"
	
	ioutil.WriteFile(listFilePath, []byte(listFileContent), 0700)
	ioutil.WriteFile(md5FilePath, []byte(md5FileContent), 0700)
	
	waitForFileChange := make(chan bool)
	waitForCommand := make(chan bool)
	
	go func(doneChan chan bool) {		
		defer func() {
			doneChan <- true
		}()

		fmt.Println("Waiting for file list to be saved...")
		err := watchFile(listFilePath)
		if err != nil {
			criticalError(err)
		}
		
		fmt.Println("File has been changed")
	}(waitForFileChange)

	go func(doneChan chan bool) {	
		defer func() {
			doneChan <- true
		}()

		err := editFile(listFilePath)
		if err != nil {
			criticalError(err)
		}
	}(waitForCommand)
	
	<- waitForCommand
	<- waitForFileChange
}