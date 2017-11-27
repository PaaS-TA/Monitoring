package layercake

import (
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/archive"
)

//go:generate counterfeiter -o fake_id/fake_id.go . ID
type ID interface {
	GraphID() string
}

//go:generate counterfeiter -o fake_cake/fake_cake.go . Cake
type Cake interface {
	DriverName() string
	Create(layerID, parentID ID, containerID string) error
	Register(img *image.Image, layer archive.ArchiveReader) error
	Get(id ID) (*image.Image, error)
	Unmount(id ID) error
	Remove(id ID) error
	Path(id ID) (string, error)
	QuotaedPath(id ID, quota int64) (string, error)
	IsLeaf(id ID) (bool, error)
	GetAllLeaves() ([]ID, error)
	All() []*image.Image
}
