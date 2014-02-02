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
	"os/signal"
	"os/user"
	"path/filepath"
	"github.com/jessevdk/go-flags"	
	"sort"
	"strings"
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

func filePathsFromListFile(filePath string) ([]string, error) {
	contentB, err := ioutil.ReadFile(filePath)
	if err != nil {
		return []string{}, err
	}
	
	var output []string
	content := string(contentB)
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.Trim(line, "\n\r\t ")
		if line == "" {
			continue
		}
		output = append(output, line)
	}
	
	return output, nil
}

func twoColumnPrint(col1 []string, col2 []string, separator string) {
	if len(col1) != len(col2) {
		panic("col1 and col2 length do not match")
	}
	
	maxColLength1 := 0
	for _, d1 := range col1 {
		if len(d1) > maxColLength1 {
			maxColLength1 = len(d1)
		}
	}
	
	for i, d1 := range col1 {
		d2 := col2[i]
		for len(d1) < maxColLength1 {
			d1 += " "
		}
		fmt.Println(d1 + separator + d2)
	}
}

func main() {
	// -----------------------------------------------------------------------------------
	// Handle SIGINT (Ctrl + C)
	// -----------------------------------------------------------------------------------
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func(){
		<-signalChan
		fmt.Println("\nOperation has been aborted.")
		// TODO: delete temp files
		os.Exit(2)
	}()
	
	// -----------------------------------------------------------------------------------
	// Parse arguments
	// -----------------------------------------------------------------------------------
	
	var opts struct {
		DryRun bool `short:"n" long:"dry-run" description:"Don't rename anything but show the operation that would have been performed."`
		Verbose bool `short:"v" long:"verbose" description:"Enable verbose output."`
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
	
	// -----------------------------------------------------------------------------------
	// Build file list
	// -----------------------------------------------------------------------------------
	
	listFileContent := ""
	baseFilename := ""
	for _, filePath := range filePaths {
		listFileContent += filePath + "\n"
		baseFilename += filePath + "|"
	}
	
	baseFilename = stringHash(baseFilename)
	listFilePath := configFolder() + "/" + baseFilename + ".files.txt"
	
	ioutil.WriteFile(listFilePath, []byte(listFileContent), 0700)
	
	// -----------------------------------------------------------------------------------
	// Watch for changes in file list
	// -----------------------------------------------------------------------------------
	
	waitForFileChange := make(chan bool)
	waitForCommand := make(chan bool)
	
	go func(doneChan chan bool) {		
		defer func() {
			doneChan <- true
		}()

		fmt.Println("Waiting for file list to be saved... (Press Ctrl + C to abort)")
		err := watchFile(listFilePath)
		if err != nil {
			criticalError(err)
		}
	}(waitForFileChange)
	
	// -----------------------------------------------------------------------------------
	// Launch text editor
	// -----------------------------------------------------------------------------------

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
	
	// -----------------------------------------------------------------------------------
	// Check that the filenames have not been changed while the list was being edited
	// -----------------------------------------------------------------------------------
	
	for _, filePath := range filePaths {
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			criticalError(errors.New("Filenames have been changed or some files have been deleted or moved while the list was being edited. To avoid any data loss, the operation has been aborted. You may resume it by running the same command."))
		}
	}

	// -----------------------------------------------------------------------------------
	// Get new filenames from list file
	// -----------------------------------------------------------------------------------
	
	newFilePaths, err := filePathsFromListFile(listFilePath)
	if err != nil {
		criticalError(err)		
	}
	
	if len(newFilePaths) != len(filePaths) {
		criticalError(errors.New(fmt.Sprintf("Number of files in list (%d) does not match original number of files (%d).", len(newFilePaths), len(filePaths))))
	}

	// -----------------------------------------------------------------------------------
	// Check for duplicates
	// -----------------------------------------------------------------------------------
	
	for i1, p1 := range newFilePaths {
		for i2, p2 := range newFilePaths {
			if i1 == i2 {
				continue
			}
			if p1 == p2 {
				criticalError(errors.New("There are duplicate filenames in the list. To avoid any data loss, the operation has been aborted. You may resume it by running the same command. The duplicate filenames are: " + p1))
			}
		}
	}	

	// -----------------------------------------------------------------------------------
	// Rename the files
	// -----------------------------------------------------------------------------------

	var dryRunCol1 []string
	var dryRunCol2 []string
	hasChanges := false
	
	for i, sourceFilePath := range filePaths {
		destFilePath := newFilePaths[i]
		
		if sourceFilePath == destFilePath {
			continue
		}
		
		hasChanges = true
		
		if opts.DryRun {
			dryRunCol1 = append(dryRunCol1, sourceFilePath)
			dryRunCol2 = append(dryRunCol2, destFilePath)
		} else {
			if opts.Verbose {
				fmt.Printf("\"%s\"  =>  \"%s\"\n", sourceFilePath, destFilePath) 
			}
			os.Rename(sourceFilePath, destFilePath)
		}
	}
	
	if opts.DryRun {
		twoColumnPrint(dryRunCol1, dryRunCol2, "  =>  ")
	}
	
	if !hasChanges && opts.Verbose {
		fmt.Println("No changes.")
	}
}