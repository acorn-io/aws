package tests

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/acorn-io/manager/tests/e2e"
	"github.com/acorn-io/manager/tests/helper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randToken(length int) (string, error) {
	randBytes := make([]byte, length)

	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randBytes), nil
}

func trimToDir(path, targetDir string) (string, error) {
	// Split the path into its components
	components := strings.Split(path, string(filepath.Separator))

	// Find the index of the target directory
	targetIndex := -1
	for i, component := range components {
		if component == targetDir {
			targetIndex = i
			break
		}
	}

	// If the target directory is not found, return an error
	if targetIndex == -1 {
		return "", fmt.Errorf("target directory '%s' not found in path", targetDir)
	}

	// Construct the new path from the beginning to the target directory
	newPath := filepath.Join(components[:targetIndex+1]...)
	return "/" + newPath, nil
}

func loadChangesetMap(path string) (map[string]string, error) {
	fullChangesets, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var obj any
	err = json.Unmarshal(fullChangesets, &obj)
	if err != nil {
		return nil, err
	}

	jsonMap, ok := obj.(map[string]interface{})
	if !ok {
		return nil, errors.New("could not cast object ot map")
	}

	result := make(map[string]string)
	for k, v := range jsonMap {
		str, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		result[k] = string(str)
	}

	return result, nil
}

func TestChangeset(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	awsDir, err := trimToDir(wd, "aws")
	if err != nil {
		t.Fatal(err)
	}

	err = filepath.Walk(awsDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if filepath.Base(path) == "golden-change-set.json" {
			changesetMap, err := loadChangesetMap(path)
			if err != nil {
				t.Fatal(err)
			}

			acornDir := filepath.Dir(path)
			if filepath.Base(acornDir) == "testdata" {
				acornDir = strings.TrimSuffix(acornDir, string(filepath.Separator)+"testdata")
			}

			for args, changeset := range changesetMap {
				// capture trick
				// TODO: undo this when we upgrade to Go 1.22
				changeset := changeset
				args := strings.Split(args, " ")

				ok := t.Run(filepath.Base(acornDir), func(t *testing.T) {
					t.Parallel()

					token, err := randToken(6)
					if err != nil {
						t.Fatal(err)
					}

					name := "changeset-check-" + token
					t.Cleanup(func() {
						helper.RunAcornCommand(t, "rm", name)
						time.Sleep(1 * time.Second)
						helper.RunAcornCommand(t, "rm", "--ignore-cleanup", name)
					})

					concreteArgs := []string{"run", "--dangerous", "--wait=false", "-n", name, acornDir, "--dryRun=true"}
					allArgs := append(concreteArgs, args...)
					allArgs = append(allArgs, "--changeset="+changeset)

					runResult := helper.RunAcornCommand(t, allArgs...)
					if runResult.ExitCode != 0 {
						t.Fatalf("Failed to run service acorn RunDir=%s ExitCode=%d ErrorLog=%s", acornDir, runResult.ExitCode, runResult.Stderr())
					}

					require.EventuallyWithT(t, func(c *assert.CollectT) {
						result := helper.RunAcornCommand(t, "ps", name)

						attemptsRegex := regexp.MustCompile(`previous (\d+) attempts`)
						matches := attemptsRegex.FindStringSubmatch(result.Stdout())

						if len(matches) > 1 {
							attempts, err := strconv.Atoi(matches[1])
							if err != nil {
								t.Error(err)
								return
							}

							if attempts > 1 {
								logsResult := helper.RunAcornCommand(t, "logs", name)
								t.Errorf("job failed %d times: %s", attempts, logsResult.Stderr())
								return
							}
						}

						assert.True(c, strings.Contains(result.Stdout(), "OK"))
					}, e2e.DefaultTimeout, e2e.NewDefaultTick, "failed to check status")
				})
				if !ok {
					return errors.New("failed to run subtest")
				}
			}
		}

		return nil
	})
	if err != nil {
		t.Error(err)
	}
}
