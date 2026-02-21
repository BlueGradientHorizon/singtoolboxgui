package main

import (
	"github.com/bluegradienthorizon/singtoolbox/testrunner"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driven/cores/stb"
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
	conf := basic_preferences.NewBasicPreferences(appId)
	// clip := clipboard.NewFyneClipboard(a)
	// fileExp := files.NewFyneExporter()
	downloader := downloader.NewHttpDownloader()
	coreAdapter := stb.NewSTBCore(testrunner.SingBoxCore)
	webServer := webserver.NewWebServer()

	// Services (Core)
	subsService := services.NewSubscriptionsService(conf, downloader /*clip, fileExp,*/)
	testService := services.NewTestService(conf, coreAdapter /*testerAdapter, webServer*/)
	webServerService := services.NewWebServerService(conf, webServer)

	// Driving adapters
	gioui := gioui.NewGUI(nil, subsService, testService, webServerService, conf /*passsing adapter - is it good/allowed/bad?*/)
	gioui.Run()
}
