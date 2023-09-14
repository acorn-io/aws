package cdk

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/acorn-io/aws/utils/cdk-runner/pkg/utils"
)

func GenerateTemplateFile(outputFile string) error {
	if err := GenerateCDKContext(); err != nil {
		return err
	}

	cmd := exec.Command("cdk", "synth", "--path-metadata", "false", "--lookups", "false")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		message := []byte(fmt.Sprintf("error running cdk synth: %v, %v", err, stderr.String()))
		return utils.WriteToTermLogAndError(message, err)
	}

	return writeCFNTemplate(out.String(), outputFile)
}

func writeCFNTemplate(content, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
