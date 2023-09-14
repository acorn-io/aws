package utils

import "os"

func WriteToTermLogAndError(message []byte, returnErr error) error {
	os.WriteFile("/dev/termination-log", message, 0644)
	return returnErr
}
