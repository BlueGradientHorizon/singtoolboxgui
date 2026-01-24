package appwidgets

import (
	"gioui.org/layout"
	"gioui.org/unit"
)

type (
	C = layout.Context
	D = layout.Dimensions
)

// DetailRow lays out two widgets in a horizontal row, with the left
// widget considered the "Primary" widget.
type DetailRow struct {
	layout.Inset
}

// Layout the DetailRow with the provided widgets.
func (d DetailRow) Layout(gtx C, primary, detail layout.Widget) D {
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Rigid(func(gtx C) D {
			return d.Inset.Layout(gtx, primary)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(15)}.Layout),
		layout.Rigid(func(gtx C) D {
			return d.Inset.Layout(gtx, detail)
		}),
	)
}
