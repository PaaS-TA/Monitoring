package rootfs_provider

import (
	"sync"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/garden-shed/repository_fetcher"
	"code.cloudfoundry.org/garden-shed/rootfs_spec"
	"code.cloudfoundry.org/lager"
)

type ContainerLayerCreator struct {
	graph         Graph
	volumeCreator VolumeCreator
	namespacer    Namespacer
	mutex         *sync.Mutex
}

func NewLayerCreator(
	graph Graph,
	volumeCreator VolumeCreator,
	namespacer Namespacer,
) *ContainerLayerCreator {
	return &ContainerLayerCreator{
		graph:         graph,
		volumeCreator: volumeCreator,
		namespacer:    namespacer,
		mutex:         &sync.Mutex{},
	}
}

func (provider *ContainerLayerCreator) Create(log lager.Logger, id string, parentImage *repository_fetcher.Image, spec rootfs_spec.Spec) (string, []string, error) {
	var err error
	var imageID layercake.ID = layercake.DockerImageID(parentImage.ImageID)

	if spec.Namespaced {
		provider.mutex.Lock()
		imageID, err = provider.namespace(log, imageID)
		provider.mutex.Unlock()
		if err != nil {
			return "", nil, err
		}
	}

	containerID := layercake.ContainerID(id)
	if err := provider.graph.Create(containerID, imageID, id); err != nil {
		return "", nil, err
	}

	var rootPath string
	if spec.QuotaSize > 0 && spec.QuotaScope == garden.DiskLimitScopeExclusive {
		rootPath, err = provider.graph.QuotaedPath(containerID, spec.QuotaSize)
	} else if spec.QuotaSize > 0 && spec.QuotaScope == garden.DiskLimitScopeTotal {
		rootPath, err = provider.graph.QuotaedPath(containerID, spec.QuotaSize-parentImage.Size)
	} else {
		rootPath, err = provider.graph.Path(containerID)
	}

	if err != nil {
		return "", nil, err
	}

	for _, v := range parentImage.Volumes {
		if err = provider.volumeCreator.Create(rootPath, v); err != nil {
			return "", nil, err
		}
	}

	return rootPath, parentImage.Env, nil
}

func (provider *ContainerLayerCreator) namespace(log lager.Logger, imageID layercake.ID) (layercake.ID, error) {
	namespacedImageID := layercake.NamespacedID(imageID, provider.namespacer.CacheKey())

	if _, err := provider.graph.Get(namespacedImageID); err != nil {
		if err := provider.createNamespacedLayer(log, namespacedImageID, imageID); err != nil {
			return nil, err
		}
	}

	return namespacedImageID, nil
}

func (provider *ContainerLayerCreator) createNamespacedLayer(log lager.Logger, id, parentId layercake.ID) error {
	var err error
	var path string
	if path, err = provider.createLayer(id, parentId); err != nil {
		return err
	}

	defer provider.unmountTranslationLayer(id)
	return provider.namespacer.Namespace(log, path)
}

func (provider *ContainerLayerCreator) unmountTranslationLayer(id layercake.ID) {
	if err := provider.graph.Unmount(id); err != nil {
		panic(err)
	}
}

func (provider *ContainerLayerCreator) createLayer(id, parentId layercake.ID) (string, error) {
	errs := func(err error) (string, error) {
		return "", err
	}

	if err := provider.graph.Create(id, parentId, ""); err != nil {
		return errs(err)
	}

	namespacedRootfs, err := provider.graph.Path(id)
	if err != nil {
		return errs(err)
	}

	return namespacedRootfs, nil
}
