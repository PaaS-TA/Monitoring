package aufs

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/lager"
)

type BackingStore struct {
	Logger   lager.Logger
	RootPath string
}

func (bm *BackingStore) Create(id string, quota int64) (string, error) {
	log := bm.Logger.Session("create", lager.Data{"id": id, "quota": quota})

	path := bm.backingStorePath(id)
	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("creating the backing store file: %s", err)
	}
	f.Close()

	if quota == 0 {
		return "", errors.New("cannot have zero sized quota")
	}

	if err := os.Truncate(path, quota); err != nil {
		return "", fmt.Errorf("truncating the file returned error: %s", err)
	}

	output, err := exec.Command("mkfs.ext4",
		"-O", "^has_journal",
		"-F", path,
	).CombinedOutput()
	if err != nil {
		log.Error("formatting-file", err, lager.Data{"path": path, "output": string(output)})
		return "", fmt.Errorf("formatting filesystem: %s", err)
	}

	return path, nil
}

func (bm *BackingStore) Delete(id string) error {
	if err := os.RemoveAll(bm.backingStorePath(id)); err != nil {
		return fmt.Errorf("deleting backing store file: %s", err)
	}

	return nil
}

func (bm *BackingStore) backingStorePath(id string) string {
	return filepath.Join(bm.RootPath, id)
}
