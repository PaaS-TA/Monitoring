package repository_fetcher

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/archive"
)

//go:generate counterfeiter -o fake_container_id_provider/FakeContainerIDProvider.go . ContainerIDProvider
type ContainerIDProvider interface {
	ProvideID(path string) layercake.ID
}

type Local struct {
	Cake              layercake.Cake
	DefaultRootFSPath string
	IDProvider        ContainerIDProvider

	mu sync.RWMutex
}

func (l *Local) Fetch(log lager.Logger, repoURL *url.URL, _, _ string, _ int64) (*Image, error) {
	log = log.Session("local-fetch", lager.Data{"path": repoURL})

	log.Info("start")
	defer log.Info("end")

	path := repoURL.Path
	if len(path) == 0 {
		path = l.DefaultRootFSPath
	}

	if len(path) == 0 {
		return nil, errors.New("RootFSPath: is a required parameter, since no default rootfs was provided to the server.")
	}

	id, err := l.fetch(log, path)
	return &Image{
		ImageID: id,
	}, err
}

func (l *Local) FetchID(log lager.Logger, repoURL *url.URL) (layercake.ID, error) {
	return l.IDProvider.ProvideID(repoURL.Path), nil
}

func (l *Local) fetch(log lager.Logger, path string) (string, error) {
	path, err := resolve(path)
	if err != nil {
		return "", err
	}

	id := l.IDProvider.ProvideID(path)

	// synchronize all downloads, we could optimize by only mutexing around each
	// particular rootfs path, but in practice importing local rootfses is decently fast,
	// and concurrently importing local rootfses is rare.
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, err := l.Cake.Get(id); err == nil {
		log.Info("using-cache", lager.Data{"graphID": id.GraphID()})
		return id.GraphID(), nil // use cache
	}

	tar, err := archive.Tar(path, archive.Uncompressed)
	if err != nil {
		return "", fmt.Errorf("repository_fetcher: fetch local rootfs: untar rootfs: %v", err)
	}
	defer tar.Close()

	if err := l.Cake.Register(&image.Image{ID: id.GraphID()}, tar); err != nil {
		return "", fmt.Errorf("repository_fetcher: fetch local rootfs: register rootfs: %v", err)
	}

	log.Info("fetched-uncached-rootfs", lager.Data{"graphID": id.GraphID()})
	return id.GraphID(), nil
}

func resolve(path string) (string, error) {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return "", fmt.Errorf("repository_fetcher: stat file: %s", err)
	}

	if (fileInfo.Mode() & os.ModeSymlink) == os.ModeSymlink {
		if path, err = os.Readlink(path); err != nil {
			return "", fmt.Errorf("repository_fetcher: read link: %s", err)
		}
	}
	return path, nil
}

type LayerIDProvider struct {
}

func (LayerIDProvider) ProvideID(path string) layercake.ID {
	path, err := resolve(path)
	if err != nil {
		return layercake.LocalImageID{
			Path:         path,
			ModifiedTime: time.Time{},
		}
	}

	info, err := os.Lstat(path)
	if err != nil {
		return layercake.LocalImageID{
			Path:         path,
			ModifiedTime: time.Time{},
		}
	}

	return layercake.LocalImageID{
		Path:         path,
		ModifiedTime: info.ModTime(),
	}
}
