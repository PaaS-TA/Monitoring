package repository_fetcher

import (
	"errors"
	"io"
	"net/url"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
	"github.com/docker/distribution"
	"github.com/docker/docker/registry"
)

//go:generate counterfeiter -o fake_lock/FakeLock.go . Lock
type Lock interface {
	Acquire(key string)
	Release(key string) error
}

// apes docker's *registry.Registry
type Registry interface {
	// v1 methods
	GetRepositoryData(repoName string) (*registry.RepositoryData, error)
	GetRemoteTags(registries []string, repository string) (map[string]string, error)
	GetRemoteHistory(imageID string, registry string) ([]string, error)
	GetRemoteImageJSON(imageID string, registry string) ([]byte, int, error)
	GetRemoteImageLayer(imageID string, registry string, size int64) (io.ReadCloser, error)
}

type RemoteFetcher interface {
	Fetch(request *FetchRequest) (*Image, error)
}

//go:generate counterfeiter . RepositoryFetcher
type RepositoryFetcher interface {
	Fetch(log lager.Logger, u *url.URL, username, password string, diskQuota int64) (*Image, error)
	FetchID(log lager.Logger, u *url.URL) (layercake.ID, error)
}

type FetchRequest struct {
	Session    *registry.Session
	Endpoint   *registry.Endpoint
	Repository distribution.Repository
	Path       string
	RemotePath string
	Tag        string
	Logger     lager.Logger
	MaxSize    int64
}

type Image struct {
	ImageID string
	Env     []string
	Volumes []string
	Size    int64
}

var ErrInvalidDockerURL = errors.New("invalid docker url")

// apes dockers registry.NewEndpoint
var RegistryNewEndpoint = registry.NewEndpoint

// apes dockers registry.NewSession
var RegistryNewSession = registry.NewSession
