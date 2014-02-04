package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"github.com/jessevdk/go-flags"
	"runtime"	
	"sort"
	"strings"
	"time"
)

const APPNAME = "massren"

type CommandLineOptions struct {
	DryRun bool `short:"n" long:"dry-run" description:"Don't rename anything but show the operation that would have been performed."`
	Verbose bool `short:"v" long:"verbose" description:"Enable verbose output."`
	Config bool `short:"c" long:"config" description:"Set a configuration value. eg. massren --config <name> [value]"`
	Undo bool `short:"u" long:"undo" description:"Undo a rename operation. eg. massren --undo [path]"`
}

func stringHash(s string) string {
	h := md5.New()
	io.WriteString(h, s)
	return fmt.Sprintf("%x", h.Sum(nil))	
}

func tempFolder() string {
	output := profileFolder() + "/temp"
	err := os.MkdirAll(output, CONFIG_PERM)
	if err != nil {
		panic(err)
	}
	return output
}

func criticalError(err error) {
	logError("%s", err)
	logInfo("Run '%s --help' for usage\n", APPNAME) 
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

func guessEditorCommand() (string, error) {
	switch runtime.GOOS {
		
		case "windows":
			
			return "notepad.exe", nil
		
		default: // assumes a POSIX system
		
			editors := []string{
				"nano",
				"vim",
				"emacs",
				"vi",
				"ed",
			}
			
			for _, editor := range editors {
				err := exec.Command("type", editor).Run()
				if err == nil {
					return editor, nil
				}
			}
	
	}
			
	return "", errors.New("could not guess editor command")
}

func editFile(filePath string) error {
	var err error
	editorCmd := config_.String("editor")
	if editorCmd == "" {
		editorCmd, err = guessEditorCommand()
		setupInfo := fmt.Sprintf("Run `%s --config editor \"name-of-editor\"` to set up the editor. eg. `%s --config editor \"vim\"`", APPNAME, APPNAME)
		if err != nil {
			criticalError(errors.New(fmt.Sprintf("No text editor defined in configuration, and could not guess a text editor.\n%s", setupInfo)))
		} else {
			logInfo("No text editor defined in configuration. Using \"%s\" as default. %s", editorCmd, setupInfo) 
		}
	}
	
	cmd := exec.Command(editorCmd, filePath)
	cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
	err = cmd.Run()

	if err != nil {
		return err
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
			if strings.Index(arg, "*") < 0 && strings.Index(arg, "?") < 0 {
				output = append(output, arg)
				continue
			}
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

func deleteTempFiles() error {	
	tempFiles, err := filepath.Glob(tempFolder() + "/*")
	if err != nil {
		return err
	}

	for _, p := range tempFiles {
		os.Remove(p)
	}
	
	return nil
}

func onExit() {
	deleteTempFiles()
	deleteOldHistoryItems(time.Now().Unix() - 60 * 60 * 24 * 7)
	profileClose()
}

func main() {
	minLogLevel_ = 1
	
	// -----------------------------------------------------------------------------------
	// Handle SIGINT (Ctrl + C)
	// -----------------------------------------------------------------------------------
	
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)
	go func() {
		<-signalChan
		logInfo("Operation has been aborted.")
		onExit()
		os.Exit(2)
	}()
	
	defer onExit()
	
	// -----------------------------------------------------------------------------------
	// Parse arguments
	// -----------------------------------------------------------------------------------
	
	var opts CommandLineOptions
	flagParser := flags.NewParser(&opts, flags.HelpFlag | flags.PassDoubleDash)
	args, err := flagParser.Parse()
	if err != nil {
		t := err.(*flags.Error).Type
		if t == flags.ErrHelp {
			flagParser.WriteHelp(os.Stdout)
			return
		} else {
			criticalError(err)
		}
	}
	
	if opts.Verbose {
		minLogLevel_ = 0
	}
	
	profileOpen()

	// -----------------------------------------------------------------------------------
	// Handle selected command
	// -----------------------------------------------------------------------------------
	
	var commandName string
	if opts.Config {
		commandName = "config"
	} else if opts.Undo {
		commandName = "undo"
	} else {
		commandName = "rename"
	}
	
	var commandErr error
	switch commandName {
		case "config": commandErr = handleConfigCommand(&opts, args)
		case "undo": commandErr = handleUndoCommand(&opts, args)
	}
	
	if commandErr != nil {
		criticalError(commandErr)		
	}
	
	if commandName != "rename" {
		return
	}
	
	filePaths, err := filePathsFromArgs(args)

	if err != nil {
		criticalError(err)
	}
	
	if len(filePaths) == 0 {
		criticalError(errors.New("no file to rename"))
	}
		
	// -----------------------------------------------------------------------------------
	// Build file list
	// -----------------------------------------------------------------------------------
	
	listFileContent := ""
	baseFilename := ""
	for _, filePath := range filePaths {
		listFileContent += filepath.Base(filePath) + "\n"
		baseFilename += filePath + "|"
	}
	
	baseFilename = stringHash(baseFilename)
	listFilePath := tempFolder() + "/" + baseFilename + ".files.txt"
	
	ioutil.WriteFile(listFilePath, []byte(listFileContent), CONFIG_PERM)
	
	// -----------------------------------------------------------------------------------
	// Watch for changes in file list
	// -----------------------------------------------------------------------------------
	
	waitForFileChange := make(chan bool)
	waitForCommand := make(chan bool)
	
	go func(doneChan chan bool) {		
		defer func() {
			doneChan <- true
		}()

		logInfo("Waiting for file list to be saved... (Press Ctrl + C to abort)")
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
	
	var sources []string
	var destinations []string
	defer func() {
		err := saveHistoryItems(sources, destinations)
		if err != nil {
			logError("Could not save history items: %s", err)
		}
	}()
	 
	for i, sourceFilePath := range filePaths {
		destFilePath := newFilePaths[i]
		
		if filepath.Base(sourceFilePath) == filepath.Base(destFilePath) {
			continue
		}
		
		destFilePath = filepath.Dir(sourceFilePath) + "/" + filepath.Base(destFilePath)
		
		hasChanges = true
		
		if opts.DryRun {
			dryRunCol1 = append(dryRunCol1, sourceFilePath)
			dryRunCol2 = append(dryRunCol2, destFilePath)
		} else {
			logDebug("\"%s\"  =>  \"%s\"", sourceFilePath, destFilePath) 
			err = os.Rename(sourceFilePath, destFilePath)
			if err != nil {
				criticalError(err)
			}
			sources = append(sources, sourceFilePath)
			destinations = append(destinations, destFilePath)
		}
	}
	
	if opts.DryRun {
		twoColumnPrint(dryRunCol1, dryRunCol2, "  =>  ")
	}
	
	if !hasChanges {
		logDebug("No changes.")
	}
}
