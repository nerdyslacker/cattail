package services

import (
	_ "embed"

	"github.com/energye/systray"
)

var (
	//go:embed trayicons/active.png
	iconActive []byte

	//go:embed trayicons/inactive.png
	iconInactive []byte
)

func trayIcon(isOnline bool) []byte {
	if isOnline {
		return iconActive
	}
	return iconInactive
}

type trayService struct {
	statusMenuItem *systray.MenuItem
	showMenuItem   *systray.MenuItem
	quitMenuItem   *systray.MenuItem
	isWindowHidden bool
}

func TrayService(isOnline bool) *trayService {
	systray.SetIcon(trayIcon(isOnline))
	systray.SetTitle("Cattail")

	status := systray.AddMenuItem("", "")
	systray.AddSeparator()
	show := systray.AddMenuItem("Show", "")
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
	systray.SetIcon(trayIcon(isOnline))
}
