package windows

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/style"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
)

type SettingsPage struct {
	router  *Router
	btnBack widget.Clickable
	list    widget.List

	checkEntries map[string]*appwidgets.PreferenceCheckStyle
	numEntries   map[string]*appwidgets.PreferenceNumericalEntryStyle
}

func NewSettingsPage(router *Router, th *material.Theme, p ports.Configuration) *SettingsPage {
	checkEntries := make(map[string]*appwidgets.PreferenceCheckStyle)
	checkEntries["dedup"] = appwidgets.PreferenceCheck(th, &widget.Bool{}, "Perform deduplication", p.DedupEnabled())
	checkEntries["batches"] = appwidgets.PreferenceCheck(th, &widget.Bool{}, "Test by batches", p.EnableBatches())
	checkEntries["autoStart"] = appwidgets.PreferenceCheck(th, &widget.Bool{}, "Auto-start web-server atfer test", p.AutoStartSrv())
	checkEntries["autoStop"] = appwidgets.PreferenceCheck(th, &widget.Bool{}, "Auto-stop web-server after first request", p.AutoStopSrv())
	checkEntries["localhostOnly"] = appwidgets.PreferenceCheck(th, &widget.Bool{}, "Web-server listens only localhost", p.SrvLocalhostOnly())

	numEntries := make(map[string]*appwidgets.PreferenceNumericalEntryStyle)
	numEntries["subDlTimeout"] = appwidgets.PreferenceNumericalEntry(th, "Subscription download timeout", "Seconds", p.SubscriptionDlTimeout())
	numEntries["rounds"] = appwidgets.PreferenceNumericalEntry(th, "Recheck rounds", "Number", p.RecheckRounds())
	numEntries["timeout"] = appwidgets.PreferenceNumericalEntry(th, "Round timeout", "Seconds", p.RoundTimeout())
	numEntries["batchSize"] = appwidgets.PreferenceNumericalEntry(th, "Batch size", "Number", p.BatchSize())
	port := appwidgets.PreferenceNumericalEntry(th, "Web-server port", "1-65535", p.SrvPort())
	port.NumericalEntry.Max = 65535
	numEntries["port"] = port

	sp := &SettingsPage{
		router:       router,
		list:         widget.List{List: layout.List{Axis: layout.Vertical}},
		checkEntries: checkEntries,
		numEntries:   numEntries,
	}
	return sp
}

func (p *SettingsPage) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.btnBack.Clicked(gtx) {
		p.router.Pop()
	}

	for _, v := range p.checkEntries {
		v.Update(gtx)
	}
	for _, v := range p.numEntries {
		v.Update(gtx)
	}

	widgets := []layout.Widget{
		p.numEntries["subDlTimeout"].Layout,
		p.checkEntries["dedup"].Layout,
		p.numEntries["rounds"].Layout,
		p.numEntries["timeout"].Layout,
		p.checkEntries["batches"].Layout,
		p.numEntries["batchSize"].Layout,
		p.checkEntries["autoStart"].Layout,
		p.numEntries["port"].Layout,
		p.checkEntries["autoStop"].Layout,
		p.checkEntries["localhostOnly"].Layout,
	}

	return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return appwidgets.MaterialButton(th, &p.btnBack, "Back", nil, "").Layout(gtx)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.List(th, &p.list).Layout(gtx, len(widgets), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.UniformInset(style.DefaultMargin).Layout(gtx, widgets[i])
				})
			}),
		)
	})
}

func (p *SettingsPage) OnPushed() {
	for _, v := range p.numEntries {
		v.SetCurrentPrefValue()
	}
}
