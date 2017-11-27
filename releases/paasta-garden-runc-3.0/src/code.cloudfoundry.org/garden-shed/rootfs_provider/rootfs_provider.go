package rootfs_provider

import "code.cloudfoundry.org/garden-shed/layercake"

type Graph interface {
	layercake.Cake
}
