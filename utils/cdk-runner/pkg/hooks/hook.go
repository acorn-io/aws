package hooks

import (
	"bytes"
	"os"
	"os/exec"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/utils"
	"github.com/sirupsen/logrus"
)

const PreChangeSetApplyHookExecutable = "/app/hooks/pre-change-set-apply"

var (
	currentTemplate = "/app/current-template.yaml"
	newTemplate     = "/app/cfn.yaml"
	changeSet       = "/app/change-set.json"
)

func PreChangeSetApplyHook(executable string) error {
	// Check exec file exists and is executable
	info, err := os.Stat(executable)
	if os.IsNotExist(err) {
		logrus.Infof("no pre-change-set-apply hook found at %s", executable)
		return nil
	} else if err != nil {
		return err
	}

	// Check if exec file is executable
	if info.Mode()&0111 == 0 {
		logrus.Infof("pre-change-set-apply hook found at %s but is not executable", executable)
		return nil
	}

	logrus.Infof("running pre-change-set-apply hook at %s", executable)
	cmd := exec.Command(executable, currentTemplate, newTemplate, changeSet)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return utils.WriteToTermLogAndError(stderr.Bytes(), err)
	}

	return nil
}
