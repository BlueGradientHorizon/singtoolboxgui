package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func ShowQuestion(a fyne.App) {
	w := a.NewWindow("Help")
	md := widget.NewRichTextFromMarkdown("# SingToolbox\n\n1. Add subscriptions.\n2. Click **Update** to fetch URIs.\n3. Click **Test** to check latency.")
	w.SetContent(container.NewScroll(md))
	w.Resize(fyne.NewSize(400, 300))
	w.Show()
}
