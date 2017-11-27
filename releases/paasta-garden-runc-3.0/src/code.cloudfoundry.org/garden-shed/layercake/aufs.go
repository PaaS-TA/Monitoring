package layercake

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/commandrunner"

	"fmt"

	"github.com/docker/docker/image"
)

const (
	metadataDirName    string = "garden-info"
	parentChildDirName string = "parent-child"
	childParentDirName string = "child-parent"
)

type AufsCake struct {
	Cake
	Runner    commandrunner.CommandRunner
	GraphRoot string
}

func (a *AufsCake) Create(childID, parentID ID, id string) error {
	if _, ok := childID.(NamespacedLayerID); !ok {
		return a.Cake.Create(childID, parentID, id)
	}

	if isAlreadyNamespaced, err := a.hasInfo(a.childParentDir(), childID); err != nil {
		return err
	} else if isAlreadyNamespaced {
		return fmt.Errorf("%s already exists", childID.GraphID())
	}

	if err := a.Cake.Create(childID, DockerImageID(""), ""); err != nil {
		return err
	}

	_, err := a.Cake.Get(childID)
	if err != nil {
		return err
	}

	sourcePath, err := a.Cake.Path(parentID)
	if err != nil {
		return err
	}
	defer a.Cake.Unmount(parentID)

	destinationPath, err := a.Cake.Path(childID)
	if err != nil {
		return err
	}

	copyCmd := fmt.Sprintf("cp -a %s/. %s", sourcePath, destinationPath)
	if err := a.Runner.Run(exec.Command("sh", "-c", copyCmd)); err != nil {
		return err
	}

	if err = a.addInfo(a.parentChildDir(), parentID.GraphID(), childID.GraphID()); err != nil {
		return err
	}

	if err = a.addInfo(a.childParentDir(), childID.GraphID(), parentID.GraphID()); err != nil {
		return err
	}

	return nil
}

func (a *AufsCake) IsLeaf(id ID) (bool, error) {
	if isDockerLeaf, err := a.Cake.IsLeaf(id); err != nil {
		return false, err
	} else if !isDockerLeaf {
		return false, nil
	}

	isParent, err := a.hasInfo(a.parentChildDir(), id)
	if err != nil {
		return false, err
	}

	return !isParent, nil
}

func (a *AufsCake) GetAllLeaves() ([]ID, error) {
	var leaves []ID

	dockerLeaves, err := a.Cake.GetAllLeaves()
	if err != nil {
		return []ID{}, err
	}

	for _, dockerLeaf := range dockerLeaves {
		isParent, err := a.hasInfo(a.parentChildDir(), dockerLeaf)
		if err != nil {
			return []ID{}, err
		}

		if !isParent {
			leaves = append(leaves, dockerLeaf)
		}
	}

	return leaves, nil
}

func (a *AufsCake) Get(id ID) (*image.Image, error) {
	img, err := a.Cake.Get(id)
	if err != nil {
		return nil, err
	}

	if img.Parent == "" {
		parentData, err := a.readInfo(a.childParentDir(), id)
		if err != nil {
			return nil, err
		}

		img.Parent = strings.TrimSpace(parentData)
	}
	return img, nil
}

func (a *AufsCake) Remove(id ID) error {
	if err := a.Cake.Remove(id); err != nil {
		return err
	}

	parentData, err := a.readInfo(a.childParentDir(), id)
	if err != nil {
		return err
	}

	parentGraphID := strings.TrimSpace(string(parentData))
	if err := os.Remove(filepath.Join(a.childParentDir(), id.GraphID())); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("layercake: Remove failed to remove file %s", err)
	}

	if err := a.removeInfo(a.parentChildDir(), parentGraphID, id.GraphID()); err != nil {
		return err
	}

	return nil
}

func (a *AufsCake) readInfo(path string, id ID) (string, error) {
	parentData, err := ioutil.ReadFile(filepath.Join(a.childParentDir(), id.GraphID()))
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(parentData), nil
}

func (a *AufsCake) removeInfo(path string, file string, content string) error {
	filePath := filepath.Join(path, file)
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	graphIDs := strings.Split(string(fileData), "\n")
	finalGraphIDs := []string{}
	for _, ID := range graphIDs {
		if ID != content && ID != "" {
			finalGraphIDs = append(finalGraphIDs, ID)
		}
	}

	if err := os.RemoveAll(filePath); err != nil {
		return err
	}

	for _, ID := range finalGraphIDs {
		if err = a.addInfo(path, file, ID); err != nil {
			return err
		}
	}
	return nil
}

func (a *AufsCake) hasInfo(path string, id ID) (bool, error) {
	if _, err := os.Stat(filepath.Join(path, id.GraphID())); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (a *AufsCake) addInfo(path string, file string, content string) error {

	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}

	handle, err := os.OpenFile(
		filepath.Join(path, file),
		os.O_CREATE|os.O_RDWR|os.O_APPEND,
		0755)
	if err != nil {
		return err
	}
	defer handle.Close()

	if _, err := fmt.Fprintln(handle, content); err != nil {
		return err
	}

	return nil
}

func (a *AufsCake) parentChildDir() string {
	return filepath.Join(a.GraphRoot, metadataDirName, parentChildDirName)
}

func (a *AufsCake) childParentDir() string {
	return filepath.Join(a.GraphRoot, metadataDirName, childParentDirName)
}
