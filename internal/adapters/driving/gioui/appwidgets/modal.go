package appwidgets

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

// Modal is a reusable wrapper around component.ModalLayer.
// It handles the click shield, background styling, and alpha fading automatically.
type Modal struct {
	layer  *component.ModalLayer
	shield widget.Clickable
}

// NewModal creates a new generic modal.
func NewModal() *Modal {
	m := component.NewModal()
	m.VisibilityAnimation.Duration = 100 * time.Millisecond
	return &Modal{
		layer: m,
	}
}

// Show triggers the modal to appear with the provided content.
// content: A function that draws the widgets inside the modal.
func (m *Modal) Show(gtx layout.Context, content func(gtx layout.Context, th *material.Theme) layout.Dimensions) {
	// We configure the internal widget logic once when Show is called.
	m.layer.Widget = func(gtx layout.Context, th *material.Theme, anim *component.VisibilityAnimation) layout.Dimensions {

		// 1. Calculate Opacity
		progress := anim.Revealed(gtx)

		// 2. Clone Theme and apply Alpha
		localTh := *th
		localTh.Palette.Bg = mulAlpha(th.Palette.Bg, progress)
		localTh.Palette.Fg = mulAlpha(th.Palette.Fg, progress)
		localTh.Palette.ContrastBg = mulAlpha(th.Palette.ContrastBg, progress)
		localTh.Palette.ContrastFg = mulAlpha(th.Palette.ContrastFg, progress)

		// 3. Center the Modal
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			// 4. The Click Shield (Prevents clicks falling through to scrim)
			return m.shield.Layout(gtx, func(gtx layout.Context) layout.Dimensions {

				// 5. Record content to measure size
				macro := op.Record(gtx.Ops)
				dims := content(gtx, &localTh)
				call := macro.Stop()

				// 6. Draw Background (Surface)
				rect := image.Rectangle{Max: dims.Size}
				// Draw rounded rectangle background using the faded Theme Background
				cl := clip.UniformRRect(rect, gtx.Dp(8)).Push(gtx.Ops)
				paint.Fill(gtx.Ops, localTh.Palette.Bg)
				cl.Pop()

				// 7. Draw the recorded content on top
				call.Add(gtx.Ops)

				return dims
			})
		})
	}

	m.layer.Appear(gtx.Now)
}

// Dismiss closes the modal.
func (m *Modal) Dismiss(gtx layout.Context) {
	m.layer.Disappear(gtx.Now)
}

// Layout draws the modal layer. This must be called last in your main layout.
func (m *Modal) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return m.layer.Layout(gtx, th)
}

// mulAlpha helper
func mulAlpha(c color.NRGBA, alpha float32) color.NRGBA {
	c.A = uint8(float32(c.A) * alpha)
	return c
}
