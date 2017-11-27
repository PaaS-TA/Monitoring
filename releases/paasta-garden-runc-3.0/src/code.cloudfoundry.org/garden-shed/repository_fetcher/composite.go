package repository_fetcher

import (
	"fmt"
	"net/url"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
)

type CompositeFetcher struct {
	// fetcher used for requests without a scheme
	LocalFetcher RepositoryFetcher

	// fetchers used for docker:// urls, depending on the version
	RemoteFetcher RepositoryFetcher
}

func (f *CompositeFetcher) Fetch(log lager.Logger, repoURL *url.URL, username, password string, diskQuota int64) (*Image, error) {
	if repoURL.Scheme == "" {
		return f.LocalFetcher.Fetch(log, repoURL, "", "", diskQuota)
	}

	return f.RemoteFetcher.Fetch(log, repoURL, username, password, diskQuota)
}

func (f *CompositeFetcher) FetchID(log lager.Logger, repoURL *url.URL) (layercake.ID, error) {
	if repoURL.Scheme == "" {
		return f.LocalFetcher.FetchID(log, repoURL)
	}

	return f.RemoteFetcher.FetchID(log, repoURL)
}

type dockerImage struct {
	layers []*dockerLayer
}

func (d dockerImage) Env() []string {
	var envs []string
	for _, l := range d.layers {
		envs = append(envs, l.env...)
	}

	return envs
}

func (d dockerImage) Vols() []string {
	var vols []string
	for _, l := range d.layers {
		vols = append(vols, l.vols...)
	}

	return vols
}

type dockerLayer struct {
	env  []string
	vols []string
	size int64
}

func FetchError(context, registry, reponame string, err error) error {
	return garden.NewServiceUnavailableError(fmt.Sprintf("repository_fetcher: %s: could not fetch image %s from registry %s: %s", context, reponame, registry, err))
}
