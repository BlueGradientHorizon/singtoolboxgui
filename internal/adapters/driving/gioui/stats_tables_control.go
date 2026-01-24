package gioui

import (
	"errors"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
)

func (g *GUI) AddStatsTables(batches int, rounds int) error {
	if batches <= 0 {
		return errors.New("AddStatsTable: batches cannot be less than or equal to 0")
	}
	if rounds <= 0 {
		return errors.New("AddStatsTable: rounds cannot be less than or equal to 0")
	}

	for range batches {
		t := appwidgets.NewStatsTable(rounds)
		g.StatsTables = append(g.StatsTables, t)
	}

	g.Window.Invalidate()
	return nil
}

func (g *GUI) CleanStatsTables() {
	g.StatsTables = nil
	g.Window.Invalidate()
}
