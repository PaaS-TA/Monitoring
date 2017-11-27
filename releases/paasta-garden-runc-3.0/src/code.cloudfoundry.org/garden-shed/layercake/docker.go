// Abstracts a layered filesystem provider, such as docker's Graph
package layercake

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/graph"
	"github.com/docker/docker/image"
	"github.com/docker/docker/pkg/archive"
)

type QuotaedDriver interface {
	graphdriver.Driver
	GetQuotaed(id, mountlabel string, quota int64) (string, error)
}

type Docker struct {
	Graph  *graph.Graph
	Driver graphdriver.Driver
}

func (d *Docker) DriverName() string {
	return d.Driver.String()
}

func (d *Docker) Create(layerID, parentID ID, containerID string) error {
	return d.Register(
		&image.Image{
			ID:        layerID.GraphID(),
			Parent:    parentID.GraphID(),
			Container: containerID,
		}, nil)
}

func (d *Docker) Register(image *image.Image, layer archive.ArchiveReader) error {
	return d.Graph.Register(&descriptor{image}, layer)
}

func (d *Docker) Get(id ID) (*image.Image, error) {
	return d.Graph.Get(id.GraphID())
}

func (d *Docker) Unmount(id ID) error {
	return d.Driver.Put(id.GraphID())
}

func (d *Docker) Remove(id ID) error {
	if err := d.Driver.Put(id.GraphID()); err != nil {
		return err
	}

	return d.Graph.Delete(id.GraphID())
}

func (d *Docker) Path(id ID) (result string, err error) {
	for i := 0; i < 5; i++ {
		result, err = d.Driver.Get(id.GraphID(), "")
		if err == nil {
			return
		}
	}
	return
}

func (d *Docker) QuotaedPath(id ID, quota int64) (string, error) {
	if d.DriverName() == "aufs" {
		return d.Driver.(QuotaedDriver).GetQuotaed(id.GraphID(), "", quota)
	} else {
		return "", errors.New("quotas are not supported for this driver")
	}
}

func (d *Docker) All() (layers []*image.Image) {
	for _, layer := range d.Graph.Map() {
		layers = append(layers, layer)
	}
	return layers
}

func (d *Docker) IsLeaf(id ID) (bool, error) {
	heads := d.Graph.Heads()
	_, ok := heads[id.GraphID()]
	return ok, nil
}

func (d *Docker) GetAllLeaves() ([]ID, error) {
	heads := d.Graph.Heads()
	var result []ID

	for head := range heads {
		result = append(result, DockerImageID(head))
	}

	return result, nil
}

type ContainerID string
type DockerImageID string

type LocalImageID struct {
	Path         string
	ModifiedTime time.Time
}

type NamespacedLayerID struct {
	LayerID  ID
	CacheKey string
}

func NamespacedID(id ID, cacheKey string) NamespacedLayerID {
	return NamespacedLayerID{id, cacheKey}
}

func (c ContainerID) GraphID() string {
	return shaID(string(c))
}

func (d DockerImageID) GraphID() string {
	return string(d)
}

func (c LocalImageID) GraphID() string {
	return shaID(fmt.Sprintf("%s-%d", c.Path, c.ModifiedTime))
}

func (n NamespacedLayerID) GraphID() string {
	return shaID(n.LayerID.GraphID() + "@" + n.CacheKey)
}

func shaID(id string) string {
	if id == "" {
		return id
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(id)))
}

type descriptor struct {
	image *image.Image
}

func (d descriptor) ID() string {
	return d.image.ID
}

func (d descriptor) Parent() string {
	return d.image.Parent
}

func (d descriptor) MarshalConfig() ([]byte, error) {
	return json.Marshal(d.image)
}
