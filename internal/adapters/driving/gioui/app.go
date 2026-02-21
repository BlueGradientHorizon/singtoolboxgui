package gioui

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/style"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/windows"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/common"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type GUI struct {
	Window           *app.Window
	Router           *windows.Router
	SubsService      ports.SubscriptionsService
	TestService      ports.TestService
	WebServerService ports.WebServerService
	Prefs            ports.Configuration

	// UI State
	Theme *material.Theme

	BtnUpdate    appwidgets.MaterialButtonStyle
	BtnSubs      appwidgets.MaterialButtonStyle
	BtnHelp      appwidgets.MaterialButtonStyle
	BtnSettings  appwidgets.MaterialButtonStyle
	BtnCopy      appwidgets.MaterialButtonStyle
	BtnExport    appwidgets.MaterialButtonStyle
	BtnWebServer appwidgets.MaterialButtonStyle
	BtnTest      appwidgets.MaterialButtonStyle

	LblFoundText      string
	LblDuplicatedText string
	LblParseErrText   string
	LblValidErrText   string
	LblWorkingText    string

	ProgressBarVal float32
	ProgressBarMax float32
	ProgressBarFmt func() string

	StatsList   widget.List
	StatsTables []*appwidgets.StatsTable

	IsValidating  bool
	TestCtx       *context.Context
	TestCtxCancel *context.CancelFunc

	// Progress bar control channels
	ProgressBarAnimEnd chan<- struct{}

	Explorer *explorer.Explorer
}

func NewGUI(
	_ any, // Placeholder for fyne.App
	ms ports.SubscriptionsService,
	ts ports.TestService,
	ws ports.WebServerService,
	p ports.Configuration,
) *GUI {
	return &GUI{
		SubsService:      ms,
		TestService:      ts,
		WebServerService: ws,
		Prefs:            p,
		ProgressBarFmt:   func() string { return "Ready" },
	}
}

func (g *GUI) Run() {
	go func() {
		g.Window = new(app.Window)
		if err := g.RunLoop(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func (g *GUI) RunLoop() error {
	g.Window.Option(app.Title("SingToolBoxGUI"), app.Size(unit.Dp(450), unit.Dp(650)))

	g.Theme = material.NewTheme()
	g.Theme.Shaper = text.NewShaper(text.WithCollection(gofont.Collection()))

	g.Router = windows.NewRouter(g.Window)

	g.Router.Register(windows.RouteMain, &DashboardPage{gui: g})
	g.Router.Register(windows.RouteHelp, windows.NewHelpPage(g.Router))
	g.Router.Register(windows.RouteSettings, windows.NewSettingsPage(g.Router, g.Theme, g.Prefs))
	g.Router.Register(windows.RouteSubs, windows.NewSubscriptionsPage(g.Router, g.Prefs))
	// g.Router.Register(windows.RouteSubEditor, windows.NewSubscriptionEditorPage(g.Router, g.Prefs))

	icHelp, _ := widget.NewIcon(icons.ActionHelp)
	icSettings, _ := widget.NewIcon(icons.ActionSettings)

	g.BtnUpdate = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "Update", nil, "")
	g.BtnSubs = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "Subscriptions", nil, "")
	g.BtnHelp = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "", icHelp, "Help")
	g.BtnSettings = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "", icSettings, "Settings")
	g.BtnCopy = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "Copy", nil, "")
	g.BtnExport = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "Export", nil, "")
	g.BtnWebServer = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "Web-Server", nil, "")
	g.BtnTest = appwidgets.MaterialButton(g.Theme, &widget.Clickable{}, "TEST", nil, "")

	g.UpdateStatsLabels()

	g.Explorer = explorer.NewExplorer(g.Window)

	var ops op.Ops
	for {
		e := g.Window.Event()
		g.Explorer.ListenEvents(e)
		switch e := e.(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			g.Router.Layout(gtx, g.Theme)
			e.Frame(gtx.Ops)
		}
	}
}

type DashboardPage struct {
	gui *GUI
}

func (p *DashboardPage) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return p.gui.LayoutDashboard(gtx)
}

func (p *DashboardPage) OnPushed() {}

func (g *GUI) LayoutDashboard(gtx layout.Context) layout.Dimensions {
	if g.BtnHelp.Button.Clicked(gtx) {
		g.Router.Push(windows.RouteHelp)
	}
	if g.BtnSettings.Button.Clicked(gtx) {
		g.Router.Push(windows.RouteSettings)
	}
	if g.BtnSubs.Button.Clicked(gtx) {
		g.Router.Push(windows.RouteSubs)
	}
	if g.BtnUpdate.Button.Clicked(gtx) {
		g.UpdateButtonAction()
	}
	if g.BtnCopy.Button.Clicked(gtx) {
		g.CopyButtonAction(gtx)
	}
	if g.BtnExport.Button.Clicked(gtx) {
		g.ExportButtonAction()
	}
	if g.BtnWebServer.Button.Clicked(gtx) {
		g.WebServerButtonAction()
	}
	if g.BtnTest.Button.Clicked(gtx) {
		g.TestButtonAction()
	}

	g.StatsList.Axis = layout.Vertical

	inset := layout.UniformInset(style.DefaultMargin)
	return inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Top Bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, layout.Spacer{}.Layout),
					layout.Rigid(g.BtnHelp.Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Rigid(g.BtnSettings.Layout),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Controls
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(1, g.BtnSubs.Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, g.BtnUpdate.Layout),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Stats Labels
			layout.Rigid(material.Body1(g.Theme, g.LblFoundText).Layout),
			layout.Rigid(material.Body1(g.Theme, g.LblDuplicatedText).Layout),
			layout.Rigid(material.Body1(g.Theme, g.LblParseErrText).Layout),
			layout.Rigid(material.Body1(g.Theme, g.LblValidErrText).Layout),
			layout.Rigid(material.Body1(g.Theme, g.LblWorkingText).Layout),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Stats Tables
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				l := material.List(g.Theme, &g.StatsList)
				l.AnchorStrategy = material.Overlay
				return l.Layout(gtx, len(g.StatsTables), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(layout.Spacer{Height: style.DefaultMargin * 2}.Layout),
						layout.Rigid(material.H6(g.Theme, fmt.Sprintf("Batch %d", i+1)).Layout),
						layout.Rigid(layout.Spacer{Height: style.DefaultMargin * 2}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return g.StatsTables[i].Layout(gtx, g.Theme)
						}),
					)
				})
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Progress Bar
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						p := material.ProgressBar(g.Theme, g.ProgressBarVal)
						return p.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, material.Caption(g.Theme, g.ProgressBarFmt()).Layout)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Action Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Horizontal}.Layout(gtx,
					layout.Flexed(1, g.BtnCopy.Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, g.BtnExport.Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, g.BtnWebServer.Layout),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),

			// Test Button
			layout.Rigid(g.BtnTest.Layout),
		)
	})
}

func (g *GUI) UpdateStatsLabels() {
	fmtStat := func(val int) string {
		if val < 0 {
			return "N/A"
		}
		return strconv.Itoa(val)
	}

	g.LblFoundText = "Profiles found total: " + fmtStat(g.Prefs.ProfilesFoundTotal().Get())
	g.LblDuplicatedText = "Profiles duplicated total: " + fmtStat(g.Prefs.ProfilesDuplicatedTotal().Get())
	g.LblParseErrText = "Parsing errors total: " + fmtStat(g.Prefs.ParsingErrorsTotal().Get())
	g.LblValidErrText = "Validation errors total: " + fmtStat(g.Prefs.ValidationErrorsTotal().Get())
	g.LblWorkingText = "Working profiles total: " + fmtStat(g.Prefs.WorkingProfilesTotal().Get())
	if g.Window != nil {
		g.Window.Invalidate()
	}
}

func (g *GUI) getUpdatingText(success int, fail int) string {
	return fmt.Sprintf("Updating (%d/%d/%d)", len(g.Prefs.Subscriptions().Get()), success, fail)
}

func (g *GUI) UpdateButtonAction() {
	go func() {
		g.BtnUpdate.Text = g.getUpdatingText(0, 0)
		g.Window.Invalidate()

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
			g.BtnUpdate.Text = g.getUpdatingText(success, fail)
			g.Window.Invalidate()
		}
		g.UpdateStatsLabels()
		g.BtnUpdate.Text = "Update"
		g.Window.Invalidate()
	}()
}

func (g *GUI) CopyButtonAction(gtx layout.Context) {
	go func() {
		p := g.SubsService.GetWorkingSubscriptionsProfiles(true)
		ps := common.NewlineJoinedString(p, func(p domain.Profile) string { return p.URI })
		gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(ps))})
	}()
}

func (g *GUI) ExportButtonAction() {
	go func() {
		fName := time.Now().Format("singtoolboxgui-20060102-150405.txt")
		f, err := g.Explorer.CreateFile(fName)
		if err != nil {
			println(err.Error())
			return
		}
		defer f.Close()
		ps := g.SubsService.ExportWorkingProfiles()
		io.Copy(f, strings.NewReader(ps))
	}()
}

func (g *GUI) WebServerButtonAction() {
	go func() {
		g.triggerWebServer()
	}()
}

func (g *GUI) triggerWebServer() {
	onStop := func() {
		g.BtnWebServer.Background = g.Theme.ContrastBg
		g.Window.Invalidate()
	}
	if !g.WebServerService.IsWebServerRunning() {
		g.BtnWebServer.Background = color.NRGBA{R: 0, G: 128, B: 0, A: 255}
		g.Window.Invalidate()
		g.WebServerService.StartWebServer(g.SubsService.ExportWorkingProfiles(), onStop)
	} else {
		g.WebServerService.StopWebServer(onStop)
	}
}

func (g *GUI) TestButtonAction() {
	go func() {
		btnTestRestore := func() {
			g.BtnTest.Text = "TEST"
			g.BtnTest.Background = g.Theme.ContrastBg
			g.Window.Invalidate()
		}
		if g.IsValidating {
			return
		}
		if g.TestCtx != nil || g.TestCtxCancel != nil {
			g.StopTest()
		} else {
			g.IsValidating = true
			g.BtnTest.Text = "VALIDATING"
			g.Window.Invalidate()
			if g.TestService.ValidateSubscriptions() == 0 {
				btnTestRestore()
				return
			} else {
				g.UpdateStatsLabels()
			}
			g.IsValidating = false
			g.BtnTest.Text = "STOP"
			g.BtnTest.Background = color.NRGBA{R: 200, A: 255}
			g.Window.Invalidate()
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
		btnTestRestore()
	}()
}
