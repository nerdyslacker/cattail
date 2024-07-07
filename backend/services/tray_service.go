package services

import (
	"context"
	_ "embed"

	"github.com/energye/systray"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	showMenuItem *systray.MenuItem
	quitMenuItem *systray.MenuItem
}

func TrayService(isOnline bool) *trayService {
	systray.SetIcon(trayIcon(isOnline))
	systray.SetTitle("Cattail")

	show := systray.AddMenuItem("Show", "")
	systray.AddSeparator()
	quit := systray.AddMenuItem("Quit", "")

	return &trayService{
		showMenuItem: show,
		quitMenuItem: quit,
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

func (ts *trayService) SetActions(ctx context.Context) {
	ts.showMenuItem.Click(func() {
		runtime.WindowShow(ctx)
	})

	ts.quitMenuItem.Click(func() {
		ts.Stop()
		runtime.Quit(ctx)
	})

	systray.SetOnClick(func(menu systray.IMenu) {
		// x, y := runtime.WindowGetPosition(ctx)
		runtime.WindowShow(ctx)
		// if x == 0 && y == 0 {
		// 	runtime.WindowShow(ctx)
		// } else {
		// 	runtime.WindowHide(ctx)
		// }
	})
}

func (ts *trayService) setStatus(isOnline bool) {
	systray.SetIcon(trayIcon(isOnline))
}
