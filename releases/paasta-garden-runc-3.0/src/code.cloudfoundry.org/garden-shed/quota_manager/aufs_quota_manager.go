package quota_manager

import (
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter . AUFSDiffPathFinder

type AUFSDiffPathFinder interface {
	GetDiffLayerPath(rootFSPath string) string
}

//go:generate counterfeiter . BaseSizer

type BaseSizer interface {
	BaseSize(logger lager.Logger, rootfsPath string) (uint64, error)
}

//go:generate counterfeiter . DiffSizer

type DiffSizer interface {
	DiffSize(logger lager.Logger, loopdevPath string) (uint64, error)
}

type AUFSQuotaManager struct {
	BaseSizer BaseSizer
	DiffSizer DiffSizer
}

func (*AUFSQuotaManager) SetLimits(logger lager.Logger, containerRootFSPath string, limits garden.DiskLimits) error {
	return nil
}

func (*AUFSQuotaManager) GetLimits(logger lager.Logger, containerRootFSPath string) (garden.DiskLimits, error) {
	return garden.DiskLimits{}, nil
}

func (a *AUFSQuotaManager) GetUsage(logger lager.Logger, containerRootFSPath string) (garden.ContainerDiskStat, error) {
	baseSize, err := a.BaseSizer.BaseSize(logger, containerRootFSPath)
	if err != nil {
		return garden.ContainerDiskStat{}, err
	}

	diffSize, err := a.DiffSizer.DiffSize(logger, containerRootFSPath)
	if err != nil {
		return garden.ContainerDiskStat{}, err
	}

	return garden.ContainerDiskStat{
		ExclusiveBytesUsed: diffSize,
		TotalBytesUsed:     diffSize + baseSize,
	}, nil
}

func (*AUFSQuotaManager) Setup() error {
	return nil
}
