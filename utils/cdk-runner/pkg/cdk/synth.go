package cdk

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
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
		os.WriteFile("/dev/termination-log", []byte(fmt.Sprintf("error running cdk synth: %v, %v", err, stderr.String())), 0644)
		return fmt.Errorf("error running cdk synth: %v, %v", err, stderr.String())
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
