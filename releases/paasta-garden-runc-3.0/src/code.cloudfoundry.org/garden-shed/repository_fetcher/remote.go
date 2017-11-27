package repository_fetcher

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/docker/image"

	"github.com/docker/distribution/digest"

	"code.cloudfoundry.org/garden-shed/distclient"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
)

type Remote struct {
	DefaultHost string
	Dial        Dialer
	Cake        layercake.Cake
	Verifier    Verifier

	FetchLock *FetchLock
}

func NewRemote(defaultHost string, cake layercake.Cake, dialer Dialer, verifier Verifier) *Remote {
	return &Remote{
		DefaultHost: defaultHost,
		Dial:        dialer,
		Cake:        cake,
		Verifier:    verifier,
		FetchLock:   NewFetchLock(),
	}
}

func (r *Remote) Fetch(log lager.Logger, u *url.URL, username, password string, diskQuota int64) (*Image, error) {
	log = log.Session("fetch", lager.Data{"url": u})

	log.Info("start")
	defer log.Info("finished")

	conn, manifest, err := r.manifest(log, u, username, password)
	if err != nil {
		return nil, err
	}

	totalImageSize := int64(0)
	for _, layer := range manifest.Layers {
		totalImageSize += layer.Image.Size
	}

	if diskQuota > 0 && totalImageSize > diskQuota {
		return nil, ErrQuotaExceeded
	}

	var env []string
	var vols []string
	for _, layer := range manifest.Layers {
		if layer.Image.Config != nil {
			env = append(env, layer.Image.Config.Env...)
			vols = append(vols, keys(layer.Image.Config.Volumes)...)
		}

		if err := r.fetchLayer(log, conn, layer); err != nil {
			return nil, err
		}
	}

	return &Image{
		ImageID: hex(manifest.Layers[len(manifest.Layers)-1].StrongID),
		Env:     env,
		Volumes: vols,
		Size:    totalImageSize,
	}, nil
}

func (r *Remote) FetchID(log lager.Logger, u *url.URL) (layercake.ID, error) {
	_, manifest, err := r.manifest(log.Session("fetch-id"), u, "", "")
	if err != nil {
		return nil, err
	}

	return layercake.DockerImageID(hex(manifest.Layers[len(manifest.Layers)-1].StrongID)), nil
}

func (r *Remote) manifest(log lager.Logger, u *url.URL, username, password string) (distclient.Conn, *distclient.Manifest, error) {
	log = log.Session("get-manifest", lager.Data{"url": u})

	log.Debug("started")
	defer log.Debug("got")

	host := u.Host
	if host == "" {
		host = r.DefaultHost
	}

	isDockerHub := host == "registry-1.docker.io"
	path := u.Path[1:] // strip off initial '/'
	isOfficialImage := strings.Index(path, "/") < 0
	if isDockerHub && isOfficialImage {
		// The Docker Hub keeps manifests of official images under library/
		path = "library/" + path
	}

	tag := u.Fragment
	if tag == "" {
		tag = "latest"
	}

	conn, err := r.Dial.Dial(log, host, path, username, password)
	if err != nil {
		return nil, nil, err
	}

	manifest, err := conn.GetManifest(log, tag)
	if err != nil {
		return nil, nil, fmt.Errorf("get manifest for tag %s on repo %s: %s", u.Fragment, u, err)
	}

	return conn, manifest, err
}

func (r *Remote) fetchLayer(log lager.Logger, conn distclient.Conn, layer distclient.Layer) error {
	log = log.Session("fetch-layer", lager.Data{"blobsum": layer.BlobSum, "id": layer.StrongID, "parent": layer.ParentStrongID})

	log.Info("start")
	defer log.Info("fetched")

	r.FetchLock.Acquire(layer.BlobSum.String())
	defer r.FetchLock.Release(layer.BlobSum.String())

	_, err := r.Cake.Get(layercake.DockerImageID(hex(layer.StrongID)))
	if err == nil {
		log.Info("got-cache")
		return nil
	}

	blob, err := conn.GetBlobReader(log, layer.BlobSum)
	if err != nil {
		return err
	}

	log.Debug("verifying")
	verifiedBlob, err := r.Verifier.Verify(blob, layer.BlobSum)
	if err != nil {
		return err
	}

	log.Debug("verified")
	defer verifiedBlob.Close()

	log.Debug("registering")
	err = r.Cake.Register(&image.Image{
		ID:     hex(layer.StrongID),
		Parent: hex(layer.ParentStrongID),
		Size:   layer.Image.Size,
	}, verifiedBlob)
	if err != nil {
		return err
	}

	return nil
}

//go:generate counterfeiter . Dialer
type Dialer interface {
	Dial(logger lager.Logger, host, repo, username, password string) (distclient.Conn, error)
}

func keys(m map[string]struct{}) (r []string) {
	for k, _ := range m {
		r = append(r, k)
	}
	return
}

func hex(d digest.Digest) string {
	if d == "" {
		return ""
	}

	return d.Hex()
}
