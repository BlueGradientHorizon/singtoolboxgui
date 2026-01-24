package appwidgets

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/widget"
)

type ScrollForwarder struct {
	widget.BaseWidget
	OnScroll func(*fyne.ScrollEvent)
}

func NewScrollForwarder(onScroll func(*fyne.ScrollEvent)) *ScrollForwarder {
	f := &ScrollForwarder{OnScroll: onScroll}
	f.ExtendBaseWidget(f)
	return f
}

func (f *ScrollForwarder) Scrolled(ev *fyne.ScrollEvent) {
	if f.OnScroll != nil {
		f.OnScroll(ev)
	}
}

func (f *ScrollForwarder) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(canvas.NewRectangle(color.Transparent))
}
