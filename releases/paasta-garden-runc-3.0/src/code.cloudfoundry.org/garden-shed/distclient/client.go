package distclient

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/docker/docker/image"

	"code.cloudfoundry.org/lager"
	"github.com/docker/distribution"
	"github.com/docker/distribution/digest"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	"golang.org/x/net/context"
)

//go:generate counterfeiter -o fake_distclient/fake_conn.go . Conn
type Conn interface {
	GetManifest(logger lager.Logger, tag string) (*Manifest, error)
	GetBlobReader(logger lager.Logger, d digest.Digest) (io.Reader, error)
}

type conn struct {
	client distribution.Repository
}

type Manifest struct {
	Layers []Layer
}

type Layer struct {
	BlobSum        digest.Digest
	StrongID       digest.Digest
	ParentStrongID digest.Digest
	Image          image.Image
}

type dialer struct {
	InsecureRegistryList InsecureRegistryList
}

func NewDialer(insecureRegistries []string) *dialer {
	return &dialer{InsecureRegistryList(insecureRegistries)}
}

func (d dialer) Dial(logger lager.Logger, host, repo, username, password string) (Conn, error) {
	host, transport, err := newTransport(logger, d.InsecureRegistryList, host, repo, username, password)
	if err != nil {
		logger.Error("failed-to-construct-transport", err)
		return nil, err
	}

	repoClient, err := client.NewRepository(context.TODO(), repo, host, transport)
	if err != nil {
		logger.Error("failed-to-construct-repository", err)
		return nil, err
	}

	return &conn{client: repoClient}, nil
}

func (r *conn) GetManifest(logger lager.Logger, tag string) (*Manifest, error) {
	manifestService, err := r.client.Manifests(context.TODO())
	if err != nil {
		logger.Error("failed-to-construct-manifest-service", err)
		return nil, err
	}

	layer, err := manifestService.GetByTag(tag)
	if err != nil {
		logger.Error("failed-to-get-by-tag", err)
		return nil, err
	}

	layers, err := toLayers(layer.FSLayers, layer.History)
	if err != nil {
		logger.Error("failed-to-get-v1-compat-layers", err)
		return nil, err
	}

	return &Manifest{Layers: layers}, nil
}

func (r *conn) GetBlobReader(logger lager.Logger, digest digest.Digest) (io.Reader, error) {
	blobStore := r.client.Blobs(context.TODO())
	return blobStore.Open(context.TODO(), digest)
}

func toLayers(fsl []manifest.FSLayer, history []manifest.History) (r []Layer, err error) {
	var parent digest.Digest
	for i := len(fsl) - 1; i >= 0; i-- {
		var img image.Image
		err := json.Unmarshal([]byte(history[i].V1Compatibility), &img)
		if err != nil {
			return nil, err
		}

		config, err := image.MakeImageConfig([]byte(history[i].V1Compatibility), fsl[i].BlobSum, parent)
		id, err := image.StrongID(config)

		r = append(r, Layer{
			BlobSum:        fsl[i].BlobSum,
			Image:          img,
			StrongID:       id,
			ParentStrongID: parent,
		})

		parent = id
	}

	return
}

func newTransport(logger lager.Logger, insecureRegistries InsecureRegistryList, host, repo, username, password string) (string, http.RoundTripper, error) {
	scheme := "https://"
	baseTransport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		DisableKeepAlives: true,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureRegistries.AllowInsecure(host),
		},
	}

	authTransport := transport.NewTransport(baseTransport)

	pingClient := &http.Client{
		Transport: authTransport,
		Timeout:   15 * time.Second,
	}

	req, err := http.NewRequest("GET", scheme+host+"/v2/", nil)
	if err != nil {
		logger.Error("failed-to-create-ping-request", err)
		return "", nil, err
	}

	challengeManager := auth.NewSimpleChallengeManager()

	resp, err := pingClient.Do(req)
	if err != nil {
		logger.Error("failed-to-ping-registry", err)

		if !insecureRegistries.AllowInsecure(host) {
			return "", nil, err
		}

		scheme = "http://"
		req, err = http.NewRequest("GET", scheme+host+"/v2/", nil)
		if err != nil {
			logger.Error("failed-to-create-http-ping-request", err)
			return "", nil, err
		}

		resp, err = pingClient.Do(req)
		if err != nil {
			return "", nil, err
		}
	} else {
		defer resp.Body.Close()

		if err := challengeManager.AddResponse(resp); err != nil {
			logger.Error("failed-to-add-response-to-challenge-manager", err)
			return "", nil, err
		}
	}

	credentialStore := dumbCredentialStore{username, password}
	tokenHandler := auth.NewTokenHandler(authTransport, credentialStore, repo, "pull")
	basicHandler := auth.NewBasicHandler(credentialStore)
	authorizer := auth.NewAuthorizer(challengeManager, tokenHandler, basicHandler)

	return scheme + host, transport.NewTransport(baseTransport, authorizer), nil
}

type dumbCredentialStore struct {
	username string
	password string
}

func (dcs dumbCredentialStore) Basic(*url.URL) (string, string) {
	return dcs.username, dcs.password
}
