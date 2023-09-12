package util

import (
	"fmt"
	"io/ioutil"
	"os/exec"
)

func RunCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	// Start the command
	if err := cmd.Start(); err != nil {
		return "", nil
	}

	// Read output from both pipes
	outputBytes, err := ioutil.ReadAll(stdout)
	if err != nil {
		panic(err)
	}
	// Convert output to strings
	output := string(outputBytes)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		panic(err)
	}

	return output, nil
}

func KillApp(appName string) error {
	cmd := exec.Command("taskkill", "/F", "/IM", appName)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error killing process:", err)
		return err
	}

	fmt.Println("Process killed")
	return nil
}
