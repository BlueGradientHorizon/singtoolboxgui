package appwidgets

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
)

type MaterialButtonStyle struct {
	Text         string
	Color        color.NRGBA
	Font         font.Font
	TextSize     unit.Sp
	Background   color.NRGBA
	CornerRadius unit.Dp
	Inset        layout.Inset
	Button       *widget.Clickable
	Shaper       *text.Shaper
	Icon         *widget.Icon
	Size         unit.Dp
	Description  string
}

func MaterialButton(th *material.Theme, button *widget.Clickable, txt string, icon *widget.Icon, description string) MaterialButtonStyle {
	b := material.Button(th, button, txt)
	ib := material.IconButton(th, button, icon, description)
	return MaterialButtonStyle{
		Text:         txt,
		Color:        b.Color,
		Font:         b.Font,
		TextSize:     b.TextSize,
		Background:   b.Background,
		CornerRadius: b.CornerRadius,
		Inset:        b.Inset,
		Button:       button,
		Shaper:       th.Shaper,
		Icon:         icon,
		Size:         ib.Size,
		Description:  description,
	}
}

func (b MaterialButtonStyle) Layout(gtx layout.Context) layout.Dimensions {
	// Measure the standard text height
	m := op.Record(gtx.Ops)
	labelDims := widget.Label{Alignment: text.Middle}.Layout(
		gtx, b.Shaper, b.Font, b.TextSize, "Aa", m.Stop(),
	)
	textHeight := labelDims.Size.Y

	// Matches vanilla material button height
	totalHeight := textHeight + gtx.Dp(b.Inset.Top) + gtx.Dp(b.Inset.Bottom)
	iconVisualSize := gtx.Dp(b.Size)

	iconWidget := func(gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			if b.Icon != nil {
				gtx.Constraints.Min = image.Point{X: iconVisualSize, Y: iconVisualSize}
				return b.Icon.Layout(gtx, b.Color)
			}
			return layout.Dimensions{}
		})
	}

	spacer := func(gtx layout.Context) layout.Dimensions {
		if b.Icon != nil {
			return layout.Spacer{Width: unit.Dp(8)}.Layout(gtx)
		}
		return layout.Dimensions{}
	}

	labelWidget := func(gtx layout.Context) layout.Dimensions {
		colMacro := op.Record(gtx.Ops)
		paint.ColorOp{Color: b.Color}.Add(gtx.Ops)
		return widget.Label{Alignment: text.Middle}.Layout(
			gtx, b.Shaper, b.Font, b.TextSize, b.Text, colMacro.Stop(),
		)
	}

	return material.ButtonLayoutStyle{
		Background:   b.Background,
		CornerRadius: b.CornerRadius,
		Button:       b.Button,
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		semantic.Button.Add(gtx.Ops)
		if d := b.Description; d != "" {
			semantic.DescriptionOp(b.Description).Add(gtx.Ops)
		}
		// Fix the button's vertical size strictly
		gtx.Constraints.Min.Y = totalHeight
		gtx.Constraints.Max.Y = totalHeight

		if b.Text == "" {
			// Icon-only button
			gtx.Constraints.Min.X = totalHeight
			gtx.Constraints.Max.X = totalHeight
			return layout.Center.Layout(gtx, iconWidget)
		}

		// Icon + text button
		horizontalInsets := layout.Inset{Left: b.Inset.Left, Right: b.Inset.Right}

		return horizontalInsets.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{
					Axis:      layout.Horizontal,
					Alignment: layout.Middle,
				}.Layout(gtx,
					layout.Rigid(iconWidget),
					layout.Rigid(spacer),
					layout.Rigid(labelWidget),
				)
			})
		})
	})
}
