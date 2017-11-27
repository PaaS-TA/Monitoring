package rootfs_provider

import (
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
)

type GetUsageFunc func(logger lager.Logger, rootfsPath string) (garden.ContainerDiskStat, error)

// MetricsAdapter implements cakeordinator.Metricser using existing quota_manager.GetUsage func
type MetricsAdapter struct {
	fn      GetUsageFunc
	id2path func(layercake.ID) string
}

func NewMetricsAdapter(fn GetUsageFunc, id2path func(layercake.ID) string) MetricsAdapter {
	return MetricsAdapter{fn, id2path}
}

func (m MetricsAdapter) Metrics(logger lager.Logger, id layercake.ID) (garden.ContainerDiskStat, error) {
	return m.fn(logger, m.id2path(id))
}
