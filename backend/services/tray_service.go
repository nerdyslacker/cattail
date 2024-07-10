package services

import (
	"bytes"
	// _ "embed"
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/energye/systray"
	"github.com/sergeymakinen/go-ico"
)

// var (
// 	//go:embed trayicons/active.png
// 	iconActive []byte

// 	//go:embed trayicons/inactive.png
// 	iconInactive []byte
// )

type trayService struct {
	statusMenuItem *systray.MenuItem
	showMenuItem   *systray.MenuItem
	quitMenuItem   *systray.MenuItem
	isWindowHidden bool
}

func TrayService(isOnline bool) *trayService {
	trayIcon, templateIcon := trayIcon(isOnline)
	systray.SetTemplateIcon(templateIcon, trayIcon)
	systray.SetIcon(trayIcon)
	systray.SetTitle("Cattail")

	show := systray.AddMenuItem("Show", "")
	systray.AddSeparator()
	status := systray.AddMenuItem("", "")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")

	return &trayService{
		statusMenuItem: status,
		showMenuItem:   show,
		quitMenuItem:   quit,
		isWindowHidden: true,
	}
}

var systrayExit = make(chan func(), 1)

func (ts *trayService) Start(onReady func()) {
	start, stop := systray.RunWithExternalLoop(onReady, nil)
	select {
	case f := <-systrayExit:
		f()
	default:
	}

	start()
	systrayExit <- stop
}

func (ts *trayService) Stop() {
	select {
	case f := <-systrayExit:
		f()
	default:
	}
}

func (ts *trayService) ToggleStatusItem(enabled bool) {
	if enabled {
		ts.statusMenuItem.Check()
		ts.statusMenuItem.SetTitle("Stop")
	} else {
		ts.statusMenuItem.Uncheck()
		ts.statusMenuItem.SetTitle("Start")
	}
}

func (ts *trayService) setStatus(isOnline bool) {
	trayIcon, templateIcon := trayIcon(isOnline)
	systray.SetTemplateIcon(templateIcon, trayIcon)
	systray.SetIcon(trayIcon)
}

func trayIcon(isOnline bool) ([]byte, []byte) {
	iconActive, iconInactive, templateIconActive, templateIconInactive, err := loadIcons()
	if err != nil {
		log.Fatalf("Failed to load icons: %v", err)
	}

	if isOnline {
		return iconActive, templateIconActive
	}

	return iconInactive, templateIconInactive
}

func loadIcons() ([]byte, []byte, []byte, []byte, error) {
	// Get the directory of the current source file
	_, filename, _, _ := runtime.Caller(0)
	sourceDir := filepath.Dir(filename)

	// Navigate to the trayIcons directory
	iconDir := filepath.Join(sourceDir, "trayicons")

	// Check if the directory exists
	if _, err := os.Stat(iconDir); os.IsNotExist(err) {
		return nil, nil, nil, nil, fmt.Errorf("icon directory does not exist: %s", iconDir)
	}

	activeIconPath := filepath.Join(iconDir, "active.ico")
	inactiveIconPath := filepath.Join(iconDir, "inactive.ico")

	activePngIconPath := filepath.Join(iconDir, "active.png")
	inactivePngIconPath := filepath.Join(iconDir, "inactive.png")

	iconActive, err := loadAndConvertIcon(activeIconPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load active icon: %w", err)
	}

	iconInactive, err := loadAndConvertIcon(inactiveIconPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load inactive icon: %w", err)
	}

	fallbackIconActive, err := loadAndConvertIcon(activePngIconPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load active icon: %w", err)
	}

	fallbackIconInactive, err := loadAndConvertIcon(inactivePngIconPath)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to load inactive icon: %w", err)
	}

	return iconActive, iconInactive, fallbackIconActive, fallbackIconInactive, nil
}

func loadAndConvertIcon(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open icon file: %w", err)
	}
	defer file.Close()

	var img image.Image

	if filepath.Ext(path) == ".ico" {
		icoFile, err := ico.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ICO file: %w", err)
		}
		// Choose the first image from the ICO file
		img = icoFile
	} else {
		img, _, err = image.Decode(file)
		if err != nil {
			return nil, fmt.Errorf("failed to decode image file: %w", err)
		}
	}

	// Convert the image to PNG format
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, fmt.Errorf("failed to encode image to PNG: %w", err)
	}

	return buf.Bytes(), nil
}
