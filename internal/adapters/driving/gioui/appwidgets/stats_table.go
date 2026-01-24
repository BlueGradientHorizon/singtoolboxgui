package appwidgets

import (
	"fmt"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type RoundStats struct {
	Running   int
	Succeeded int
	Failed    int
	Total     int
}

type StatsTableColumn int

const (
	ColTotal StatsTableColumn = iota
	ColRunning
	ColFailed
	ColSucceeded
	ColCount
)

func (c StatsTableColumn) String() string {
	return []string{"Total", "Running", "Failed", "Succeeded"}[c]
}

type StatsTable struct {
	Stats []RoundStats
	grid  component.GridState
}

func NewStatsTable(rounds int) *StatsTable {
	return &StatsTable{
		Stats: make([]RoundStats, rounds),
	}
}

func (t *StatsTable) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	borderColor := color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	headerBg := color.NRGBA{R: 45, G: 45, B: 45, A: 255}
	headerFg := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	totalCols := int(ColCount) + 1

	ts := component.Table(th, &t.grid)
	ts.AnchorStrategy = material.Overlay

	rowHeight := gtx.Dp(unit.Dp(30))

	// .Disabled() is for scrolling pass-through
	return ts.Layout(gtx.Disabled(), len(t.Stats), totalCols,
		func(axis layout.Axis, index, constraint int) int {
			if axis == layout.Horizontal {
				if index == 0 {
					return rowHeight
				}
				remainingWidth := constraint - rowHeight
				remainingCols := totalCols - 1

				if remainingCols < 1 {
					return 0
				}
				return remainingWidth / remainingCols
			}
			return rowHeight
		},
		func(gtx layout.Context, col int) layout.Dimensions {
			return widget.Border{
				Color: borderColor,
				Width: unit.Dp(0.5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				paint.FillShape(gtx.Ops, headerBg, clip.Rect{Max: gtx.Constraints.Max}.Op())
				if col == 0 {
					return layout.Dimensions{Size: gtx.Constraints.Max}
				}
				lbl := material.Body1(th, StatsTableColumn(col-1).String())
				lbl.Color = headerFg
				lbl.Alignment = text.Middle
				return layout.Center.Layout(gtx, lbl.Layout)
			})
		},
		func(gtx layout.Context, row, col int) layout.Dimensions {
			return widget.Border{
				Color: borderColor,
				Width: unit.Dp(0.5),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				if col == 0 {
					paint.FillShape(gtx.Ops, headerBg, clip.Rect{Max: gtx.Constraints.Max}.Op())
					lbl := material.Body1(th, fmt.Sprintf("%d", row+1))
					lbl.Color = headerFg
					lbl.Alignment = text.Middle
					return layout.Center.Layout(gtx, lbl.Layout)
				}

				var txt string
				switch StatsTableColumn(col - 1) {
				case ColTotal:
					txt = fmt.Sprintf("%d", t.Stats[row].Total)
				case ColRunning:
					txt = fmt.Sprintf("%d", t.Stats[row].Running)
				case ColFailed:
					txt = fmt.Sprintf("%d", t.Stats[row].Failed)
				case ColSucceeded:
					txt = fmt.Sprintf("%d", t.Stats[row].Succeeded)
				}

				lbl := material.Body1(th, txt)
				lbl.Alignment = text.Middle
				return layout.Center.Layout(gtx, lbl.Layout)
			})
		},
	)
}

func (t *StatsTable) UpdateRow(row int, stats RoundStats) {
	if row >= 0 && row < len(t.Stats) {
		t.Stats[row] = stats
	}
}
