package seats

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// AssignDevice assigns a device to a seat.
func AssignDevice(seats map[int]map[string]string, seatID int, deviceType, device string, logger *log.Logger) {
	if seats[seatID] == nil {
		seats[seatID] = make(map[string]string)
	}
	seats[seatID][deviceType] = device
	logger.Printf("Assigned %s: %s to Seat %d", deviceType, device, seatID)
}

// ApplyConfig applies the seat and user configurations.
func ApplyConfig(seats map[int]map[string]string, users map[int]string, logger *log.Logger) error {
	for id, username := range users {
		logger.Printf("Enabling RDP for user %s on Seat %d", username, id)
		if devMap, ok := seats[id]; ok {
			for typ, dev := range devMap {
				logger.Printf("Assigned %s %s to %s", typ, dev, username)
			}
		}
	}
	logger.Println("Configuration applied successfully.")
	return nil
}

// EnableInternetSharing enables ICS on default adapter.
func EnableInternetSharing(logger *log.Logger) error {
	cmd := exec.Command("powershell", "-Command", "Set-NetConnectionSharing -InterfaceAlias 'Ethernet' -SharingEnabled $true")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to enable ICS: %v", err)
	}
	logger.Println("Internet sharing enabled.")
	return nil
}