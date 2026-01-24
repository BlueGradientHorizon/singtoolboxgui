package main

import (
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driven/network/downloader"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driven/network/webserver"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driven/storage/basic_preferences"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/services"
)

func main() {
	appId := "com.bghorizon.singtoolboxgui"
	// a := app.NewWithID(appId)

	// Driven adapters
	// conf := fyne_preferences.NewFynePreferences(a.Preferences())
	conf := basic_preferences.NewBasicPreferences(appId)
	// clip := clipboard.NewFyneClipboard(a)
	// fileExp := files.NewFyneExporter()
	downloader := downloader.NewHttpDownloader()
	// testerAdapter := tester.NewSingBoxTester()
	webServer := webserver.NewWebServer()

	// Services (Core)
	subsService := services.NewSubscriptionsService(conf, downloader /*clip, fileExp,*/)
	testService := services.NewTestService(conf /*testerAdapter, webServer*/)
	webServerService := services.NewWebServerService(conf, webServer)

	// Driving adapters
	// go func() {
	// fyneUi := fyne_ui.NewGUI(a, subsService, testService, webServerService, conf)
	// fyneUi.Run()
	// }()
	gioui := gioui.NewGUI(nil, subsService, testService, webServerService, conf /*passsing adapter - is it good/allowed/bad?*/)
	gioui.Run()
}
