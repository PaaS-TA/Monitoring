package rootfs_provider

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/lager"
)

//go:generate counterfeiter -o fake_namespacer/fake_namespacer.go . Namespacer
type Namespacer interface {
	CacheKey() string
	Namespace(log lager.Logger, rootfsPath string) error
}

//go:generate counterfeiter -o fake_translator/fake_translator.go . Translator
type Translator interface {
	CacheKey() string
	Translate(path string, info os.FileInfo, err error) error
}

type UidNamespacer struct {
	Translator Translator
}

func (n *UidNamespacer) Namespace(log lager.Logger, rootfsPath string) error {
	log = log.Session("namespace-rootfs", lager.Data{
		"path": rootfsPath,
	})

	log.Info("namespace")

	if err := filepath.Walk(rootfsPath, n.Translator.Translate); err != nil {
		log.Error("walk-failed", err)
	}

	log.Info("namespaced")

	return nil
}

func (n *UidNamespacer) CacheKey() string {
	return n.Translator.CacheKey()
}
