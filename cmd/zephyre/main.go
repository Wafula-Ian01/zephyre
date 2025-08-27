package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/zephyre/internal/devices"
	"github.com/zephyre/internal/seats"
	"github.com/zephyre/internal/users"
)

const (
	MaxSeats = 10
	LogFile  = "zephyre.log"
)

var (
	appInstance fyne.App
	mainWindow  fyne.Window
	devicesMap  = make(map[string][]string) // "monitors", "keyboards", "mice"
	seatsMap    = make(map[int]map[string]string)
	usersMap    = make(map[int]string)
	logger      *log.Logger
)

func init() {
	file, err := os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	multi := io.MultiWriter(os.Stdout, file)
	logger = log.New(multi, "Zephyre: ", log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {
	appInstance = app.NewWithID("org.humanitarian.zephyre.multiseat")
	appInstance.SetIcon(theme.ComputerIcon())
	appInstance.Settings().SetTheme(theme.LightTheme())

	mainWindow = appInstance.NewWindow("Zephyre Multiseat - For Education & Aid")
	mainWindow.Resize(fyne.NewSize(900, 700))

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Scan", theme.SearchIcon(), scanTab()),
		container.NewTabItemWithIcon("Assign", theme.ViewFullScreenIcon(), assignTab()),
		container.NewTabItemWithIcon("Users", theme.AccountIcon(), usersTab()),
		container.NewTabItemWithIcon("Apply", theme.SettingsIcon(), applyTab()),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	mainWindow.SetContent(container.NewBorder(
		widget.NewLabelWithStyle("Setup Multi-User PC for Underprivileged Access", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		nil, nil, nil, tabs,
	))
	mainWindow.ShowAndRun()
}

func scanTab() fyne.CanvasObject {
	progress := widget.NewProgressBarInfinite()
	progress.Hide()
	scanButton := widget.NewButtonWithIcon("Scan Hardware", theme.SearchIcon(), func() {
		progress.Show()
		err := devices.ScanDevices(devicesMap, logger)
		if err != nil {
			logger.Printf("Scan error: %v", err)
		}
		progress.Hide()
	})
	deviceList := widget.NewRichTextFromMarkdown("**Devices:**\n\nScan to list monitors and USB peripherals.")
	updateButton := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		deviceList.ParseMarkdown(formatDevices())
	})

	return container.NewVBox(
		widget.NewLabel("Step 1: Detect monitors, keyboards, and mice for multi-user setup."),
		scanButton,
		progress,
		updateButton,
		deviceList,
	)
}

func assignTab() fyne.CanvasObject {
	seatOptions := generateSeatOptions()
	seatSelect := widget.NewSelect(seatOptions, nil)
	deviceTypeSelect := widget.NewSelect([]string{"monitor", "keyboard", "mouse"}, nil)
	deviceSelect := widget.NewSelect([]string{}, nil)
	assignButton := widget.NewButtonWithIcon("Assign", theme.ConfirmIcon(), func() {
		seatID := parseSeatID(seatSelect.Selected)
		deviceType := deviceTypeSelect.Selected
		device := deviceSelect.Selected
		if device != "" {
			seats.AssignDevice(seatsMap, seatID, deviceType, device, logger)
		}
	})
	assignmentList := widget.NewRichTextFromMarkdown("**Assignments:**\n\nAssign devices to seats.")
	updateAssignments := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		assignmentList.ParseMarkdown(formatAssignments())
		updateDeviceSelect(deviceSelect, deviceTypeSelect.Selected)
	})
	deviceTypeSelect.OnChanged = func(_ string) {
		updateDeviceSelect(deviceSelect, deviceTypeSelect.Selected)
	}

	return container.NewVBox(
		widget.NewLabel("Step 2: Assign hardware to up to 10 independent seats via GUI."),
		container.NewGridWithColumns(3, seatSelect, deviceTypeSelect, deviceSelect),
		assignButton,
		updateAssignments,
		assignmentList,
	)
}

func usersTab() fyne.CanvasObject {
	seatOptions := generateSeatOptions()
	seatSelect := widget.NewSelect(seatOptions, nil)
	usernameEntry := widget.NewEntry()
	usernameEntry.SetPlaceHolder("Enter username for seat")
	createUserButton := widget.NewButtonWithIcon("Create User", theme.AccountIcon(), func() {
		seatID := parseSeatID(seatSelect.Selected)
		username := usernameEntry.Text
		if username != "" {
			err := users.CreateWindowsUser(username, logger)
			if err != nil {
				logger.Printf("Error creating user %s: %v", username, err)
				return
			}
			usersMap[seatID] = username
			logger.Printf("Created user %s for Seat %d", username, seatID)
			usernameEntry.SetText("")
		}
	})
	userList := widget.NewRichTextFromMarkdown("**Users:**\n\nCreate users for seats.")

	return container.NewVBox(
		widget.NewLabel("Step 3: Create user accounts for concurrent access."),
		container.NewGridWithColumns(2, seatSelect, usernameEntry),
		createUserButton,
		userList,
	)
}

func applyTab() fyne.CanvasObject {
	statusLabel := widget.NewLabel("Status: Ready.")
	applyButton := widget.NewButtonWithIcon("Apply & Launch", theme.MediaPlayIcon(), func() {
		statusLabel.SetText("Applying...")
		err := seats.ApplyConfig(seatsMap, usersMap, logger)
		if err != nil {
			logger.Printf("Apply error: %v", err)
			statusLabel.SetText("Error applying config.")
			return
		}
		statusLabel.SetText("Config applied! Sessions ready for concurrent use.")
	})
	shareInternetButton := widget.NewButtonWithIcon("Enable Internet Sharing", theme.RadioButtonIcon(), func() {
		err := seats.EnableInternetSharing(logger)
		if err != nil {
			logger.Printf("Internet sharing error: %v", err)
		} else {
			statusLabel.SetText("Internet sharing enabled via ICS.")
		}
	})

	return container.NewVBox(
		widget.NewLabel("Step 4: Apply assignments and enable shared internet."),
		applyButton,
		shareInternetButton,
		statusLabel,
		widget.NewLabel("Note: Concurrency via RDP sessions; devices redirect in RDP clients."),
	)
}

// Helper functions
func generateSeatOptions() []string {
	var opts []string
	for i := 1; i <= MaxSeats; i++ {
		opts = append(opts, fmt.Sprintf("Seat %d", i))
	}
	return opts
}

func parseSeatID(s string) int {
	var id int
	fmt.Sscanf(s, "Seat %d", &id)
	return id
}

func formatDevices() string {
	var sb strings.Builder
	for cat, devs := range devicesMap {
		sb.WriteString(fmt.Sprintf("**%s:**\n", strings.Title(cat)))
		for i, d := range devs {
			sb.WriteString(fmt.Sprintf("- %d: %s\n", i+1, d))
		}
	}
	return sb.String()
}

func formatAssignments() string {
	var sb strings.Builder
	for id, devMap := range seatsMap {
		sb.WriteString(fmt.Sprintf("**Seat %d:**\n", id))
		for typ, dev := range devMap {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", typ, dev))
		}
	}
	return sb.String()
}

func updateDeviceSelect(sel *widget.Select, devType string) {
	var opts []string
	if devs, ok := devicesMap[devType+"s"]; ok {
		for i, d := range devs {
			opts = append(opts, fmt.Sprintf("#%d: %s", i+1, d))
		}
	}
	sel.Options = opts
	sel.Refresh()
}
