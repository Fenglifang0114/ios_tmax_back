package svc

import (
	"os/exec"
	"runtime"
	"strings"
	l "tmaxsrv/log"

	"go.bug.st/serial"
)

func getPortsList() ([]string, error) {
	if runtime.GOOS == "android" {
		// Android cannot scan /dev directly due to SELinux.
		// It causes avc: denied logs which crashes the WiFi driver (WifiVendorHal).
		// Android uses Flutter's usb_serial plugin instead.
		return []string{}, nil
	}

	if arch := runtime.GOARCH; arch != "arm" {

		ports, err := serial.GetPortsList()
		if err != nil {
			l.Log.Error(err)
		}
		if len(ports) == 0 {
			l.Log.Errorln("No serial ports found!")
		}
		return ports, err
	}

	// for ARM32 only
	command := "ls -l /dev/tty* | grep 'dialout' | rev | cut -d \" \" -f1 | rev"

	cmd := exec.Command("bash", "-c", command)

	output, err := cmd.Output()
	if err != nil {
		l.Log.Error("Failed to execute command:", err)
		return nil, err
	}

	outputStr := string(output)
	ports := strings.Split(strings.TrimSpace(outputStr), "\n")
	return ports, nil
}
