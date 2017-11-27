package rootfs_provider

import (
	"net/url"
	"sync"

	specs "github.com/opencontainers/runtime-spec/specs-go"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/repository_fetcher"
	"code.cloudfoundry.org/garden-shed/rootfs_spec"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . LayerCreator
type LayerCreator interface {
	Create(log lager.Logger, id string, parentImage *repository_fetcher.Image, spec rootfs_spec.Spec) (string, []string, error)
}

//go:generate counterfeiter . RepositoryFetcher
type RepositoryFetcher interface {
	Fetch(log lager.Logger, rootfs *url.URL, username, password string, diskQuota int64) (*repository_fetcher.Image, error)
}

//go:generate counterfeiter . GCer
type GCer interface {
	GC(log lager.Logger, cake layercake.Cake) error
}

//go:generate counterfeiter . Metricser
type Metricser interface {
	Metrics(logger lager.Logger, id layercake.ID) (garden.ContainerDiskStat, error)
}

// CakeOrdinator manages a cake, fetching layers as neccesary
type CakeOrdinator struct {
	mu sync.RWMutex

	cake         layercake.Cake
	fetcher      RepositoryFetcher
	layerCreator LayerCreator
	metrics      Metricser
	gc           GCer
}

// New creates a new cake-ordinator, there should only be one CakeOrdinator
// for a particular cake.
func NewCakeOrdinator(cake layercake.Cake, fetcher RepositoryFetcher, layerCreator LayerCreator, metrics Metricser, gc GCer) *CakeOrdinator {
	return &CakeOrdinator{
		cake:         cake,
		fetcher:      fetcher,
		layerCreator: layerCreator,
		metrics:      metrics,
		gc:           gc,
	}
}

func (c *CakeOrdinator) Create(logger lager.Logger, id string, spec rootfs_spec.Spec) (specs.Spec, error) {
	logger = logger.Session("create", lager.Data{"id": id})
	logger.Info("start")
	c.mu.RLock()
	defer func() {
		c.mu.RUnlock()
		logger.Info("finished")
	}()
	logger.Info("lock-acquired")

	fetcherDiskQuota := spec.QuotaSize
	if spec.QuotaScope == garden.DiskLimitScopeExclusive {
		fetcherDiskQuota = 0
	}

	image, err := c.fetcher.Fetch(logger, spec.RootFS, spec.Username, spec.Password, fetcherDiskQuota)
	if err != nil {
		return specs.Spec{}, err
	}

	rootFS, env, err := c.layerCreator.Create(logger, id, image, spec)
	if err != nil {
		return specs.Spec{}, err
	}
	return specs.Spec{
		Root:    &specs.Root{Path: rootFS},
		Process: &specs.Process{Env: env},
	}, nil
}

func (c *CakeOrdinator) Metrics(logger lager.Logger, id string, _ bool) (garden.ContainerDiskStat, error) {
	logger = logger.Session("metrics", lager.Data{"id": id})
	logger.Info("start")
	defer logger.Info("finished")

	cid := layercake.ContainerID(id)
	return c.metrics.Metrics(logger, cid)
}

func (c *CakeOrdinator) Destroy(logger lager.Logger, id string) error {
	logger = logger.Session("destroy", lager.Data{"id": id})
	logger.Info("start")
	defer logger.Info("finished")

	cid := layercake.ContainerID(id)
	if _, err := c.cake.Get(cid); err != nil {
		logger.Info("layer-already-deleted-skipping", lager.Data{"id": id, "graphID": cid, "error": err.Error()})
		return nil
	}

	return c.cake.Remove(cid)
}

func (c *CakeOrdinator) GC(logger lager.Logger) error {
	logger = logger.Session("gc")
	logger.Info("start")
	c.mu.Lock()
	defer func() {
		c.mu.Unlock()
		logger.Info("finished")
	}()
	logger.Info("lock-acquired")

	return c.gc.GC(logger, c.cake)
}
