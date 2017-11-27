package cleaner

import (
	"code.cloudfoundry.org/garden-shed/layercake"
	"code.cloudfoundry.org/lager"
)

type threshold int64
type disabled bool

func NewThreshold(limit int64) Threshold {
	if limit < 0 {
		return disabled(false)
	}

	return threshold(limit)
}

func (t threshold) Exceeded(log lager.Logger, cake layercake.Cake) bool {
	log = log.Session("threshold", lager.Data{"limit": t})
	log.Info("start")

	var size int64
	for _, layer := range cake.All() {
		size += layer.Size

		log.Info("layer", lager.Data{"size": layer.Size, "total": size})
		if size > int64(t) {
			log.Info("finish", lager.Data{"exceeded": true})
			return true
		}
	}

	log.Info("finish", lager.Data{"exceeded": false})
	return false
}

func (disabled) Exceeded(log lager.Logger, cake layercake.Cake) bool {
	return false
}
