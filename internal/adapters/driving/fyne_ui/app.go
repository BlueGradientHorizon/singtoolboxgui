package fyne_ui

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/fyne_ui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/fyne_ui/windows"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/common"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type GUI struct {
	App        fyne.App
	MainWindow fyne.Window

	SubsService      ports.SubscriptionsService
	TestService      ports.TestService
	WebServerService ports.WebServerService
	Prefs            ports.Configuration

	BtnUpdate            *widget.Button
	BtnTest              *widget.Button
	BtnWebServer         *widget.Button
	LblFound             *widget.Label
	LblParseErr          *widget.Label
	LblValidErr          *widget.Label
	LblWorking           *widget.Label
	ProgressBar          *widget.ProgressBar
	StatsTablesContainer *fyne.Container
	StatsTablesScroll    *container.Scroll
	StatsTables          []*appwidgets.StatsTable

	TestCtx       *context.Context
	TestCtxCancel *context.CancelFunc
}

func NewGUI(
	a fyne.App,
	ms ports.SubscriptionsService,
	ts ports.TestService,
	ws ports.WebServerService,
	p ports.Configuration,
) *GUI {
	return &GUI{
		App:              a,
		SubsService:      ms,
		TestService:      ts,
		WebServerService: ws,
		Prefs:            p,
	}
}

func (g *GUI) Run() {
	g.MainWindow = g.App.NewWindow("SingToolBoxGUI")

	// Top Bar
	topBar := container.NewHBox(
		layout.NewSpacer(),
		widget.NewButtonWithIcon("", theme.QuestionIcon(), func() { windows.ShowQuestion(g.App) }),
		widget.NewButtonWithIcon("", theme.SettingsIcon(), func() { windows.ShowSettings(g.App, g.Prefs) }),
	)

	// Subscriptions & Update
	g.BtnUpdate = widget.NewButton("Update", g.UpdateButtonAction)
	btnSubs := widget.NewButton("Subscriptions", func() { windows.ShowSubscriptions(g.App, g.Prefs) })

	// Stats Labels
	g.LblFound = widget.NewLabel("")
	g.LblParseErr = widget.NewLabel("")
	g.LblValidErr = widget.NewLabel("")
	g.LblWorking = widget.NewLabel("")
	g.UpdateStatsLabels()

	g.ProgressBar = widget.NewProgressBar()
	g.ProgressBar.TextFormatter = g.ReadyProgressBarFormatter

	// Action Buttons
	btnCopy := widget.NewButton("Copy", g.CopyButtonAction)
	btnExport := widget.NewButton("Export", g.ExportButtonAction)
	g.BtnWebServer = widget.NewButton("Web-Server", g.WebServerButtonAction)

	g.BtnTest = widget.NewButton("TEST", g.TestButtonAction)
	g.BtnTest.Importance = widget.HighImportance

	g.StatsTablesContainer = container.NewVBox()
	statsTablesContainerBackground := canvas.NewRectangle(theme.Color(theme.ColorNameHover))
	statsTablesContainerStack := container.NewStack(statsTablesContainerBackground, g.StatsTablesContainer)
	g.StatsTablesScroll = container.NewVScroll(statsTablesContainerStack)

	content := container.NewBorder(
		container.NewVBox(
			topBar,
			container.NewGridWithColumns(2, btnSubs, g.BtnUpdate),
			g.LblFound, g.LblParseErr, g.LblValidErr, g.LblWorking,
		),
		container.NewVBox(
			g.ProgressBar,
			container.NewGridWithColumns(3, btnCopy, btnExport, g.BtnWebServer),
			g.BtnTest,
		),
		nil,
		nil,
		g.StatsTablesScroll,
	)

	g.MainWindow.SetContent(content)
	g.MainWindow.Resize(fyne.NewSize(450, 650))
	g.MainWindow.ShowAndRun()
}

func (g *GUI) UpdateStatsLabels() {
	fmtStat := func(val int) string {
		if val < 0 {
			return "N/A"
		}
		return strconv.Itoa(val)
	}

	fyne.Do(func() {
		g.LblFound.SetText("Profiles found total: " + fmtStat(g.Prefs.ProfilesFoundTotal().Get()))
		g.LblParseErr.SetText("Parsing errors total: " + fmtStat(g.Prefs.ParsingErrorsTotal().Get()))
		g.LblValidErr.SetText("Validation errors total: " + fmtStat(g.Prefs.ValidationErrorsTotal().Get()))
		g.LblWorking.SetText("Working profiles total: " + fmtStat(g.Prefs.WorkingProfilesTotal().Get()))
	})
}

func (g *GUI) getUpdatingText(success int, fail int) string {
	return fmt.Sprintf("Updating (%d/%d/%d)", len(g.Prefs.Subscriptions().Get()), success, fail)
}

func (g *GUI) UpdateButtonAction() {
	go func() {
		fyne.DoAndWait(func() {
			g.BtnUpdate.OnTapped = func() {}
			g.BtnUpdate.SetText(g.getUpdatingText(0, 0))
		})
		updateChan := make(chan domain.DownloadSubscriptionResult, len(g.Prefs.Subscriptions().Get()))
		defer close(updateChan)
		go g.SubsService.DownloadSubscriptions(time.Duration(g.Prefs.SubscriptionDlTimeout().Get()), updateChan)
		success := 0
		fail := 0
		for range len(g.Prefs.Subscriptions().Get()) {
			d := <-updateChan
			if d.Success {
				success++
			} else {
				fail++
			}
			fyne.Do(func() {
				g.BtnUpdate.SetText(g.getUpdatingText(success, fail))
			})
		}
		g.UpdateStatsLabels()
		fyne.Do(func() {
			g.BtnUpdate.OnTapped = g.UpdateButtonAction
			g.BtnUpdate.SetText("Update")
		})
	}()
}

func (g *GUI) CopyButtonAction() {
	go func() {
		p := g.SubsService.GetWorkingSubscriptionsProfiles(true)
		ps := common.NewlineJoinedString(p, func(p domain.Profile) string { return p.URI })
		g.App.Clipboard().SetContent(ps)
	}()
}

func (g *GUI) ExportButtonAction() {
	go func() {
		w := g.MainWindow
		dialog.ShowFileSave(func(uri fyne.URIWriteCloser, err error) {
			if err != nil {
				dialog.ShowError(err, w)
				return
			}

			if uri == nil {
				return
			}
			defer uri.Close()

			ps := g.SubsService.ExportWorkingProfiles()
			_, writeErr := io.Copy(uri, strings.NewReader(ps))
			if writeErr != nil {
				dialog.ShowError(writeErr, w)
				return
			}
		}, w)
	}()
}

func (g *GUI) WebServerButtonAction() {
	go func() {
		g.triggerWebServer()
	}()
}

func (g *GUI) triggerWebServer() {
	onStop := func() {
		fyne.Do(func() {
			g.BtnWebServer.Importance = widget.MediumImportance
			g.BtnWebServer.Refresh()
		})
	}
	if !g.WebServerService.IsWebServerRunning() {
		g.WebServerService.StartWebServer(g.SubsService.ExportWorkingProfiles(), nil)
		fyne.Do(func() {
			g.BtnWebServer.Importance = widget.HighImportance
			g.BtnWebServer.Refresh()
		})
	} else {
		g.WebServerService.StopWebServer(nil)
		onStop()
	}
}

func (g *GUI) TestButtonAction() {
	go func() {
		if g.TestCtx != nil && g.TestCtxCancel != nil {
			g.StopTest()
		} else {
			g.StartTest()
			g.StopTest()
			g.UpdateStatsLabels()
			if g.WebServerService.IsWebServerRunning() {
				g.triggerWebServer()
				g.triggerWebServer()
			} else if g.Prefs.AutoStartSrv().Get() {
				g.WebServerButtonAction()
			}
		}
	}()
}
