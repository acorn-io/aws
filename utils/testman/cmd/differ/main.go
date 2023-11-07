package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func fatalPrint(logLine string) {
	fmt.Println(logLine)
	os.Exit(1)
}

func mustReadFile(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		fatalPrint(fmt.Sprintf("failed to read file %s: %s", path, err.Error()))
	}

	return data
}

// we'll normalize the given JSON data by:
// 1) parse to map[string]any
// 2) filter unwanted fields (dynamic, non-critical data)
// 3) convert back to JSON
//
// this should allow us to accurately compare changesets
func normalizeJSON(data []byte) ([]byte, error) {
	var parsedJSON []map[string]interface{}

	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	for i, child := range parsedJSON {
		filterKey(child, "LogicalResourceId")
		parsedJSON[i] = child
	}

	return json.Marshal(parsedJSON)
}

// recursively filter the given key from the given map
func filterKey(m map[string]interface{}, keyToRemove string) {
	for k, v := range m {
		if k == keyToRemove {
			delete(m, k)
		}
		if nestedMap, ok := v.(map[string]any); ok {
			filterKey(nestedMap, keyToRemove)
		}
		if nestedSlice, ok := v.([]any); ok {
			for _, item := range nestedSlice {
				if nestedMap, ok := item.(map[string]any); ok {
					filterKey(nestedMap, keyToRemove)
				}
			}
		}
	}
}

func main() {
	if len(os.Args) != 4 {
		fatalPrint("unexpected args")
	}

	originalGoldenChangeset := os.Getenv("TESTCASE")
	goldenChangeset, err := normalizeJSON([]byte(originalGoldenChangeset))
	if err != nil {
		fatalPrint(fmt.Sprintf("failed to normalize the golden changeset (%s): %s", originalGoldenChangeset, err.Error()))
	}

	originalNewChangeset := mustReadFile(os.Args[3])
	newChangeset, err := normalizeJSON(originalNewChangeset)
	if err != nil {
		fatalPrint(fmt.Sprintf("failed to normalize the new changeset (%s): %s", originalNewChangeset, err.Error()))
	}

	if !bytes.Equal(goldenChangeset, newChangeset) {
		fatalPrint("new changeset does not match the golden changeset: " + string(newChangeset))
	}
}
