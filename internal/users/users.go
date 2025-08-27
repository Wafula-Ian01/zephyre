package users

import (
	"fmt"
	"log"
	"os/exec"
	"syscall"
)

// CreateWindowsUser creates a new Windows user and adds to RDP group.
func CreateWindowsUser(username string, logger *log.Logger) error {
	// Create user (default password "password"; secure in prod)
	cmd := exec.Command("net", "user", username, "password", "/add")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create user: %v, output: %s", err, output)
	}
	// Add to Remote Desktop Users
	cmd = exec.Command("net", "localgroup", "Remote Desktop Users", username, "/add")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to add to RDP group: %v, output: %s", err, output)
	}
	logger.Printf("User %s created and added to RDP group.", username)
	return nil
}