// +build linux freebsd

package trash

import (
	"os"
	"os/exec"
)

var isAvailable_ int = -1
var toolName_ string

// Tells whether it is possible to move a file to the trash
func IsAvailable() bool {
	if isAvailable_ < 0 {
		toolName_ = ""
		isAvailable_ = 0

		candidates := []string{
			"gvfs-trash",
			"trash",
		}

		for _, candidate := range candidates {
			err := exec.Command("type", candidate).Run()
			ok := false
			if err == nil {
				ok = true
			} else {
				err = exec.Command("sh", "-c", "type "+candidate).Run()
				if err == nil {
					ok = true
				}
			}

			if ok {
				toolName_ = candidate
				isAvailable_ = 1
				return true
			}
		}

		return false
	} else if isAvailable_ == 1 {
		return true
	}

	return false
}

// Move the given file to the trash
// filePath must be an absolute path
func MoveToTrash(filePath string) (string, error) {
	if IsAvailable() {
		err := exec.Command(toolName_, filePath).Run()
		return "", err
	}

	os.Remove(filePath)
	return "", nil
}
