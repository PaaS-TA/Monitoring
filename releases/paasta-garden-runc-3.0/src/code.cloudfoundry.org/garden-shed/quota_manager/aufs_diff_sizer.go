package quota_manager

import (
	"fmt"
	"os"
	"os/exec"

	"code.cloudfoundry.org/lager"
)

type AUFSDiffSizer struct {
	AUFSDiffPathFinder AUFSDiffPathFinder
}

func (a *AUFSDiffSizer) DiffSize(logger lager.Logger, containerRootFSPath string) (uint64, error) {
	_, err := os.Stat(containerRootFSPath)
	if os.IsNotExist(err) {
		return 0, fmt.Errorf("get usage: %s", err)
	}

	log := logger.Session("diff-size", lager.Data{"path": containerRootFSPath})
	log.Debug("start")

	command := fmt.Sprintf("df -B 1 %s | tail -n1 | awk -v N=3 '{print $N}'", a.AUFSDiffPathFinder.GetDiffLayerPath((containerRootFSPath)))
	outbytes, err := exec.Command("sh", "-c", command).CombinedOutput()
	if err != nil {
		log.Error("df-failed", err)
		return 0, fmt.Errorf("get usage: df: %s, %s", err, string(outbytes))
	}

	var bytesUsed uint64
	if _, err := fmt.Sscanf(string(outbytes), "%d", &bytesUsed); err != nil {
		log.Error("scanf-failed", err, lager.Data{"out": string(outbytes)})
		return 0, nil
	}

	log.Debug("finished", lager.Data{"bytes": bytesUsed})
	return bytesUsed, nil
}
