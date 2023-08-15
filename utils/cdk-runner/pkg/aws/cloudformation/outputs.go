package cloudformation

import (
	"encoding/json"
	"os"
)

func WriteOutputsToFile(c *Client, stackName, filename string) error {
	stack, err := GetStack(c, stackName)
	if err != nil {
		return err
	}
	outputs := stack.Current.Outputs

	jsonData, err := json.MarshalIndent(outputs, "", "    ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}
