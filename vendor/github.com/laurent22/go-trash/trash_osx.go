// +build darwin

package trash

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Adapted from https://github.com/morgant/tools-osx/blob/master/src/trash
func haveScriptableFinder() (bool, error) {
	// Get current user
	user, err := user.Current()
	if err != nil {
		return false, err
	}

	// Get processes for current user
	cmd := exec.Command("ps", "-u", user.Username)
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	// Find Finder process ID, if it is running
	finderPid := 0
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Index(line, "CoreServices/Finder.app") >= 0 {
			splitted := strings.Split(line, " ")
			index := 0
			for _, token := range splitted {
				if token == " " || token == "" {
					continue
				}
				index++
				if index == 2 {
					finderPid, err = strconv.Atoi(token)
					if err != nil {
						return false, err
					}
					break
				}
			}
		}
	}

	if finderPid <= 0 {
		return false, errors.New("could not find Finder process ID")
	}

	// TODO: test with screen
	if os.Getenv("STY") != "" {
		return false, errors.New("currently running in screen")
	}

	return true, nil
}

// filePath must be an absolute path
func pathVolume(filePath string) string {
	pieces := strings.Split(filePath[1:], "/")
	if len(pieces) <= 2 {
		return ""
	}
	if pieces[0] != "Volumes" {
		return ""
	}
	volumeName := pieces[1]
	cmd := exec.Command("readlink", "/Volumes/"+volumeName)
	output, _ := cmd.Output()
	if strings.Trim(string(output), " \t\r\n") == "/" {
		return ""
	}
	return volumeName
}

func fileTrashPath(filePath string) (string, error) {
	volumeName := pathVolume(filePath)
	trashPath := ""
	if volumeName != "" {
		trashPath = fmt.Sprintf("/Volumes/%s/.Trashes/%d", volumeName, os.Getuid())
	} else {
		user, err := user.Current()
		if err != nil {
			return "", err
		}
		trashPath = fmt.Sprintf("/Users/%s/.Trash", user.Username)
	}
	return trashPath, nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}

// Tells whether it is possible to move a file to the trash
func IsAvailable() bool {
	return true
}

// Move the given file to the trash
// filePath must be an absolute path
func MoveToTrash(filePath string) (string, error) {
	if !fileExists(filePath) {
		return "", errors.New("file does not exist or is not accessible")
	}

	ok, err := haveScriptableFinder()

	if ok {
		// Do this in a loop because Finder sometime randomly fails with this error:
		//     29:106: execution error: Finder got an error: Handler can’t handle objects of this class. (-10010)
		// Repeating the operation usually fixes the issue.
		maxLoop := 3
		for i := 0; i < maxLoop; i++ {
			time.Sleep(time.Duration(i*500) * time.Millisecond)

			cmd := exec.Command("/usr/bin/osascript", "-e", "tell application \"Finder\" to delete POSIX file \""+filePath+"\"")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout
			var stderr bytes.Buffer
			cmd.Stderr = &stderr
			err := cmd.Run()

			if err != nil {
				err = errors.New(fmt.Sprintf("%s: %s %s", err, stdout.String(), stderr.String()))
			}

			if stderr.Len() > 0 {
				err = errors.New(fmt.Sprintf("%s, %s", stdout.String(), stderr.String()))
			}

			if err != nil {
				if i >= maxLoop-1 {
					return "", err
				} else {
					continue
				}
			}

			break
		}
	} else {
		return "", errors.New(fmt.Sprintf("scriptable Finder not available: %s", err))

		// TODO: maybe based on https://github.com/morgant/tools-osx/blob/master/src/trash, move
		// the file to trash manually. Problem is that it won't be possible to restore the files
		// directly from the trash.

		// volumeName := pathVolume(filePath)
		// trashPath := ""
		// if volumeName != "" {
		// 	trashPath = fmt.Sprintf("/Volumes/%s/.Trashes/%d", volumeName, os.Getuid())
		// } else {
		// 	user, err := user.Current()
		// 	if err != nil {
		// 		return err
		// 	}
		// 	trashPath = fmt.Sprintf("/Users/%s/.Trash", user.Username)
		// }
		// err = os.MkdirAll(trashPath, 0700)
		// if err != nil {
		// 	return err
		// }
	}

	return "", nil

	// Code below is not working well

	trashPath, err := fileTrashPath(filePath)
	if err != nil {
		return "", err
	}

	filename := filepath.Base(filePath)
	ext := filepath.Ext(filePath)
	filenameNoExt := filename[0 : len(filename)-len(ext)]

	possibleFiles, err := filepath.Glob(trashPath + "/" + filenameNoExt + " ??.??.??" + ext)
	if err != nil {
		return "", err
	}

	latestFile := ""
	var latestTime int64
	for _, f := range possibleFiles {
		fileInfo, err := os.Stat(f)
		if err != nil {
			continue
		}
		modTime := fileInfo.ModTime().UnixNano()
		if modTime > latestTime {
			latestTime = modTime
			latestFile = f
		}
	}

	if latestFile == "" {
		return "", errors.New("could not find path of file in trash")
	}

	return latestFile, nil
}
