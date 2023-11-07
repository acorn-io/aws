package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/acorn"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/aws/cloudformation"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/cdk"
	"github.com/acorn-io/aws/utils/cdk-runner/pkg/utils"
	_ "github.com/acorn-io/baaah/pkg/logrus"
	"github.com/sirupsen/logrus"
)

const (
	CloudformationOutputFile = "outputs.json"
	AcornRenderExecutable    = "./scripts/service.sh"
)

func applyCfnTemplateFile(inputFile, stackName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	client, err := cloudformation.NewClient(ctx)
	if err != nil {
		return err
	}

	go acorn.StartEventWatcher(ctx, stackName)

	templateBytes, err := os.ReadFile(inputFile)
	if err != nil {
		return err
	}

	if err := cloudformation.DeployStack(client, stackName, string(templateBytes)); err != nil {
		return err
	}

	if err := cloudformation.WriteOutputsToFile(client, stackName, CloudformationOutputFile); err != nil {
		return err
	}
	return runServiceAcornRenderExec(AcornRenderExecutable)
}

func deleteStack(stackName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()

	client, err := cloudformation.NewClient(ctx)
	if err != nil {
		return err
	}

	go acorn.StartEventWatcher(ctx, stackName)

	return cloudformation.Delete(client, stackName)
}

func runServiceAcornRenderExec(executable string) error {
	cmd := exec.Command(executable)

	var out bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stderr = &stderr
	cmd.Stdout = &out
	return cmd.Run()

}

func waitForStackToFinishTransition(stackName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
	defer cancel()
	client, err := cloudformation.NewClient(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("stack %s did not finish transitioning after 60 minutes", stackName)
		default:
			t, status, err := cloudformation.StackOperationInProgress(client, stackName)
			if err != nil {
				return err
			}
			if t {
				logrus.Infof("Waiting: stack %s is transitioning with status %s", stackName, status)
				time.Sleep(time.Second * 30)
				continue
			}
			return nil
		}
	}
}

func main() {
	stackName := os.Getenv("ACORN_EXTERNAL_ID")
	event := os.Getenv("ACORN_EVENT")

	if err := waitForStackToFinishTransition(stackName); err != nil {
		logrus.Fatal(err)
	}

	if event == "create" || event == "update" {
		if err := cdk.GenerateTemplateFile("cfn.yaml"); err != nil {
			logrus.Fatal(err)
		}

		if err := applyCfnTemplateFile("cfn.yaml", stackName); err != nil {
			logrus.Fatal(err)
		}
	}

	if event == "delete" {
		if err := deleteStack(stackName); err != nil {
			logrus.Fatal(utils.WriteToTermLogAndError([]byte(err.Error()), err))
		}
	}
}
