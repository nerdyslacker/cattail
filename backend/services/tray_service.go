package services

import (
	_ "embed"
	"runtime"

	"cattail/backend/utils/trayicons"

	"github.com/energye/systray"
)

var (
	//go:embed trayicons/active.png
	iconActive []byte

	//go:embed trayicons/inactive.png
	iconInactive []byte
)

type trayService struct {
	statusMenuItem *systray.MenuItem
	showMenuItem   *systray.MenuItem
	quitMenuItem   *systray.MenuItem
	isWindowHidden bool
}

func TrayService(isOnline bool) *trayService {
	trayIcon := trayIcon(isOnline)
	systray.SetTemplateIcon(trayIcon, trayIcon)
	systray.SetIcon(trayIcon)
	systray.SetTitle("Cattail")
	systray.SetTooltip("Cattail")

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
	trayIcon := trayIcon(isOnline)
	systray.SetTemplateIcon(trayIcon, trayIcon)
	systray.SetIcon(trayIcon)
}

func trayIcon(isOnline bool) []byte {
	if runtime.GOOS == "windows" {
		iconActive = trayicons.Active
		iconInactive = trayicons.Inactive
	}

	if isOnline {
		return iconActive
	}

	return iconInactive
}
