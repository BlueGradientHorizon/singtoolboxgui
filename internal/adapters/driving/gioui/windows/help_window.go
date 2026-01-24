package windows

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/style"
)

type HelpPage struct {
	router  *Router
	btnBack widget.Clickable
}

func NewHelpPage(router *Router) *HelpPage {
	return &HelpPage{
		router: router,
	}
}

func (p *HelpPage) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.btnBack.Clicked(gtx) {
		p.router.Pop()
	}

	// Gio doesn't have a built-in Markdown renderer in core,
	// plain text for this example.
	txt := "SingToolbox\n\n1. Add subscriptions.\n2. Click Update to fetch URIs.\n3. Click Test to check latency."

	return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return appwidgets.MaterialButton(th, &p.btnBack, "Back", nil, "").Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return material.Body1(th, txt).Layout(gtx)
			}),
		)
	})
}

func (p *HelpPage) OnPushed() {}
