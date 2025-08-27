package devices

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32/v2"
)

// ScanDevices populates the devices map with monitors and USB peripherals.
func ScanDevices(devices map[string][]string, logger *log.Logger) error {
	err := scanMonitors(devices)
	if err != nil {
		return err
	}
	err = scanUSB(devices)
	if err != nil {
		return err
	}
	logger.Println("Hardware scan completed.")
	return nil
}

func scanMonitors(devices map[string][]string) error {
	var monitors []string
	callback := func(hmon w32.HMONITOR, hdc w32.HDC, rect *w32.RECT, lparam w32.LPARAM) uintptr {
		var info w32.MONITORINFO
		info.CbSize = uint32(unsafe.Sizeof(info))
		if w32.GetMonitorInfo(hmon, &info) {
			monitors = append(monitors, fmt.Sprintf("Monitor at (%d,%d)-(%d,%d)", info.RcMonitor.Left, info.RcMonitor.Top, info.RcMonitor.Right, info.RcMonitor.Bottom))
		}
		return 1 // Continue enumeration
	}
	cbPtr := syscall.NewCallback(callback)
	w32.EnumDisplayMonitors(0, nil, cbPtr, 0)
	devices["monitors"] = monitors
	return nil
}

func scanUSB(devices map[string][]string) error {
	cmd := exec.Command("powershell", "-Command", "Get-PnpDevice | Where-Object {$_.Class -eq 'HIDClass' -or $_.Class -eq 'Keyboard' -or $_.Class -eq 'Mouse'} | Select FriendlyName, InstanceId")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("USB scan failed: %v", err)
	}
	lines := strings.Split(string(bytes.TrimSpace(out)), "\n")
	var allUSB []string
	for i, line := range lines {
		if i > 0 && strings.TrimSpace(line) != "" {
			allUSB = append(allUSB, line)
		}
	}
	for _, u := range allUSB {
		fields := strings.Fields(u)
		if len(fields) > 0 {
			name := strings.Join(fields[:len(fields)-1], " ")
			lowerName := strings.ToLower(name)
			if strings.Contains(lowerName, "keyboard") {
				devices["keyboards"] = append(devices["keyboards"], name)
			} else if strings.Contains(lowerName, "mouse") {
				devices["mice"] = append(devices["mice"], name)
			}
		}
	}
	return nil
}
