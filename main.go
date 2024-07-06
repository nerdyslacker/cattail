package main

import (
	"cattail/backend/consts"
	"cattail/backend/services"
	"context"
	"embed"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	runtime2 "github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed build/appicon.png
var icon []byte

var version = "0.0.0"

func main() {
	// Create an instance of the app structure
	tailSvc := services.TailScaleService()
	sysSvc := services.System()
	prefSvc := services.Preferences()
	prefSvc.SetAppVersion(version)
	windowWidth, windowHeight, maximised := prefSvc.GetWindowSize()
	windowStartState := options.Normal
	if maximised {
		windowStartState = options.Maximised
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:                    "Cattail",
		Width:                    windowWidth,
		Height:                   windowHeight,
		MinWidth:                 consts.DEFAULT_WINDOW_WIDTH,
		MinHeight:                consts.DEFAULT_WINDOW_HEIGHT,
		MaxWidth:                 consts.DEFAULT_WINDOW_WIDTH,
		MaxHeight:                consts.DEFAULT_WINDOW_HEIGHT,
		WindowStartState:         windowStartState,
		Frameless:                runtime.GOOS != "darwin",
		EnableDefaultContextMenu: true,
		HideWindowOnClose:        true,
		DisableResize:            true,
		// StartHidden:              true,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: options.NewRGBA(27, 38, 54, 0),
		OnStartup: func(ctx context.Context) {
			sysSvc.Start(ctx, version)
			tailSvc.Startup(ctx)
		},
		OnDomReady: func(ctx context.Context) {
			x, y := prefSvc.GetWindowPosition(ctx)
			runtime2.WindowSetPosition(ctx, x, y)
			runtime2.WindowShow(ctx)
		},
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			x, y := runtime2.WindowGetPosition(ctx)
			prefSvc.SaveWindowPosition(x, y)
			return false
		},
		Bind: []interface{}{
			sysSvc,
			prefSvc,
			tailSvc,
		},
		Mac: &mac.Options{
			TitleBar: mac.TitleBarHiddenInset(),
			About: &mac.AboutInfo{
				Title:   "Cattail " + version,
				Message: "",
				Icon:    icon,
			},
			WebviewIsTransparent: false,
			WindowIsTranslucent:  true,
		},
		Windows: &windows.Options{
			WebviewIsTransparent:              true,
			WindowIsTranslucent:               true,
			DisableFramelessWindowDecorations: true,
		},
		Linux: &linux.Options{
			ProgramName:         "Cattail",
			Icon:                icon,
			WebviewGpuPolicy:    linux.WebviewGpuPolicyOnDemand,
			WindowIsTranslucent: true,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
