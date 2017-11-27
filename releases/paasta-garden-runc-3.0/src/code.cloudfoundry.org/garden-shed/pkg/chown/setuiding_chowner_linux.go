package chown

import (
	"os"
)

func Chown(path string, uid, gid int) error {
	origMode, err := os.Lstat(path)
	if err != nil {
		return err
	}

	err = os.Lchown(path, uid, gid)
	if err != nil {
		return err
	}

	if origMode.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	err = os.Chmod(path, origMode.Mode())
	if err != nil {
		return err
	}

	return nil
}
