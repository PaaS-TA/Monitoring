package aufs

import (
	"fmt"
	"os/exec"

	"github.com/docker/docker/daemon/graphdriver/aufs"
)

func Unmount(path string) error {
	if !isMountPoint(path) {
		return nil
	}

	if err := aufs.Unmount(path); err != nil {
		return err
	}

	if isMountPoint(path) {
		return fmt.Errorf("still a mountpoint")
	}

	return nil
}

func isMountPoint(path string) bool {
	err := exec.Command("mountpoint", path).Run()
	if err != nil {
		// if it's not a mountpoint then this is fine
		return false
	}

	return true
}
