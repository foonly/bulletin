package main

import (
	"context"
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

//go:embed frontend/public/favicon.png
var icon []byte

func main() {
	app := NewApp()

	trayMenu := menu.NewMenu()
	trayMenu.Append(menu.Text("Open Bulletin", nil, func(_ *menu.CallbackData) {
		runtime.WindowShow(app.ctx)
	}))
	trayMenu.Append(menu.Separator())
	trayMenu.Append(menu.Text("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		runtime.Quit(app.ctx)
	}))

	err := wails.Run(&options.App{
		Title:  "Bulletin",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
		OnBeforeClose: func(ctx context.Context) bool {
			runtime.WindowHide(ctx)
			return true
		},
		// In Wails v2.13+, for Linux, assigning a Menu and setting an Icon
		// in linux.Options often triggers the SNI (tray) integration automatically.
		Menu: trayMenu,
		Linux: &linux.Options{
			Icon:        icon,
			ProgramName: "bulletin",
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}
