package hooks

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/utils"
	"github.com/sirupsen/logrus"
)

const (
	PreChangeSetApplyHookExecutable = "/app/hooks/pre-change-set-apply"
	DryRunHookExecutable            = "/app/hooks/dry-run"
)

var (
	currentTemplate = "/app/current-template.yaml"
	newTemplate     = "/app/cfn.yaml"
	changeSet       = "/app/change-set.json"
)

// RunChangesetHook runs the given executable as a hook that receives the currentTemplate, newTemplate, and changeset as args
// may return an error
func RunChangesetHook(executable string) error {
	// Check exec file exists and is executable
	info, err := os.Stat(executable)
	if os.IsNotExist(err) {
		logrus.Infof("no %s hook found at %s", filepath.Base(executable), executable)
		return nil
	} else if err != nil {
		return err
	}

	// Check if exec file is executable
	if info.Mode()&0111 == 0 {
		logrus.Infof("%s hook found at %s but is not executable", filepath.Base(executable), executable)
		return nil
	}

	logrus.Infof("running %s hook at %s", filepath.Base(executable), executable)
	cmd := exec.Command(executable, currentTemplate, newTemplate, changeSet)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	defer func() {
		logrus.Info(out.String())
		logrus.Info(stderr.String())
	}()

	if err := cmd.Run(); err != nil {
		return utils.WriteToTermLogAndError(stderr.Bytes(), err)
	}

	return nil
}
