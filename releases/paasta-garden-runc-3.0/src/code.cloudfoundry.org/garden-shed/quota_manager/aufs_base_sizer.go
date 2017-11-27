package quota_manager

import (
	"fmt"
	"path"

	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
)

type AUFSBaseSizer struct {
	cake layercake.Cake
}

func NewAUFSBaseSizer(cake layercake.Cake) *AUFSBaseSizer {
	return &AUFSBaseSizer{cake: cake}
}

func (a *AUFSBaseSizer) BaseSize(logger lager.Logger, containerRootFSPath string) (uint64, error) {
	var size uint64
	graphID := path.Base(containerRootFSPath)

	for graphID != "" {
		img, err := a.cake.Get(layercake.DockerImageID(graphID))
		if err != nil {
			return 0, fmt.Errorf("base-size %s: %s", graphID, err)
		}

		logger.Debug("base-size", lager.Data{
			"layer":      graphID,
			"size":       img.Size,
			"total-size": size,
		})

		size += uint64(img.Size)
		graphID = img.Parent
	}

	return size, nil
}
