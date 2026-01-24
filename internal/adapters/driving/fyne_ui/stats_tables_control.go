package fyne_ui

import (
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/fyne_ui/appwidgets"
)

func (g *GUI) AddStatsTables(batches int, rounds int) error {
	if batches <= 0 {
		return errors.New("AddStatsTable: batches cannot be less than or equal to 0")
	}
	if rounds <= 0 {
		return errors.New("AddStatsTable: rounds cannot be less than or equal to 0")
	}

	for bI := range batches {

		t := appwidgets.NewStatsTable(rounds)

		g.StatsTables = append(g.StatsTables, t)

		labelText := fmt.Sprintf("Batch %d/%d", bI+1, batches)
		label := widget.NewLabelWithStyle(labelText, fyne.TextAlignLeading, fyne.TextStyle{})
		label.SizeName = fyne.ThemeSizeName(theme.SizeNameSubHeadingText)

		overlay := appwidgets.NewScrollForwarder(func(ev *fyne.ScrollEvent) {
			g.StatsTablesScroll.Scrolled(ev)
		})
		tableStack := container.NewStack(t, overlay)

		fyne.DoAndWait(func() {
			g.StatsTablesContainer.Add(label)
			g.StatsTablesContainer.Add(tableStack)
			g.StatsTablesContainer.Refresh()
		})
	}

	return nil
}

func (g *GUI) CleanStatsTables() {
	g.StatsTables = nil
	fyne.DoAndWait(func() { g.StatsTablesContainer.RemoveAll() })
}
