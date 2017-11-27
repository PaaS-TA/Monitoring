package aufs

import (
	"fmt"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
	"github.com/docker/docker/daemon/graphdriver"
)

type UnmountFunc func(target string) error

//go:generate counterfeiter . GraphDriver
type GraphDriver interface {
	graphdriver.Driver
}

//go:generate counterfeiter . LoopMounter
type LoopMounter interface {
	MountFile(filePath, destPath string) error
	Unmount(path string) error
}

//go:generate counterfeiter . BackingStoreMgr
type BackingStoreMgr interface {
	Create(id string, quota int64) (string, error)
	Delete(id string) error
}

//go:generate counterfeiter . Retrier
type Retrier interface {
	Run(work func() error) error
}

type QuotaedDriver struct {
	GraphDriver
	Unmount         UnmountFunc
	BackingStoreMgr BackingStoreMgr
	LoopMounter     LoopMounter
	Retrier         Retrier
	RootPath        string
	Logger          lager.Logger
}

func (a *QuotaedDriver) GetQuotaed(id, mountlabel string, quota int64) (string, error) {
	path := a.makeDiffPath(id)
	log := a.Logger.Session("get-quotaed", lager.Data{"id": id, "mountlabel": mountlabel, "quota": quota, "path": path})

	bsPath, err := a.BackingStoreMgr.Create(id, quota)
	if err != nil {
		return "", fmt.Errorf("creating backingstore file: %s", err)
	}

	if err := a.LoopMounter.MountFile(bsPath, path); err != nil {
		if err2 := a.BackingStoreMgr.Delete(id); err2 != nil {
			log.Error("cleaning-backing-store", err2)
		}

		return "", fmt.Errorf("mounting file: %s", err)
	}

	mntPath, err := a.GraphDriver.Get(id, mountlabel)
	if err != nil {
		if err2 := a.LoopMounter.Unmount(path); err2 != nil {
			log.Error("unmounting-loop-device", err2)
		}
		if err2 := a.BackingStoreMgr.Delete(id); err2 != nil {
			log.Error("cleaning-backing-store", err2)
		}

		return "", fmt.Errorf("getting mountpath: %s", err)
	}

	return mntPath, nil
}

func (a *QuotaedDriver) Put(id string) error {
	mntPath := a.makeMntPath(id)
	diffPath := a.makeDiffPath(id)

	a.GraphDriver.Put(id)

	if err := a.Retrier.Run(func() error {
		return a.Unmount(mntPath)
	}); err != nil {
		return err
	}

	if err := a.LoopMounter.Unmount(diffPath); err != nil {
		return fmt.Errorf("unmounting the loop device: %s", err)
	}

	if err := a.BackingStoreMgr.Delete(id); err != nil {
		return fmt.Errorf("removing the backing store: %s", err)
	}

	return nil
}

func (a *QuotaedDriver) makeMntPath(id string) string {
	return filepath.Join(a.RootPath, "aufs", "mnt", id)
}

func (a *QuotaedDriver) makeDiffPath(id string) string {
	return filepath.Join(a.RootPath, "aufs", "diff", id)
}

func (a *QuotaedDriver) GetDiffLayerPath(rootFSPath string) string {
	return strings.Replace(rootFSPath, "/mnt/", "/diff/", 1)
}

func (a *QuotaedDriver) GetMntPath(id layercake.ID) string {
	return a.makeMntPath(id.GraphID())
}
