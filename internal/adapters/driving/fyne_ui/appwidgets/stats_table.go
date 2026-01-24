package appwidgets

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/driver/mobile"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

type roundStats struct {
	Running   binding.Int
	Succeeded binding.Int
	Failed    binding.Int
	Total     binding.Int
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
	widget.Table
	Stats     []roundStats
	colWidths map[int]float32
}

func NewStatsTable(rounds int) *StatsTable {
	t := &StatsTable{}
	t.colWidths = make(map[int]float32)
	t.Stats = make([]roundStats, rounds)

	for i := range rounds {
		t.Stats[i].Running = binding.NewInt()
		t.Stats[i].Succeeded = binding.NewInt()
		t.Stats[i].Failed = binding.NewInt()
		t.Stats[i].Total = binding.NewInt()
	}

	updateColumnsWidths := func(id widget.TableCellID, template fyne.CanvasObject) {
		w := template.MinSize().Width + theme.Padding()*2
		if w > t.colWidths[id.Col] {
			t.colWidths[id.Col] = w
			t.SetColumnWidth(id.Col, w)
		}
	}

	t.ShowHeaderColumn = true
	t.ShowHeaderRow = true
	t.Table.Length = func() (rows int, cols int) { return rounds, int(ColCount) }
	t.Table.CreateCell = func() fyne.CanvasObject { return widget.NewLabel("") }
	t.Table.UpdateCell = func(id widget.TableCellID, o fyne.CanvasObject) {
		l := o.(*widget.Label)
		l.TextStyle = fyne.TextStyle{}
		l.Unbind()

		var b binding.Int
		switch StatsTableColumn(id.Col) {
		case ColRunning:
			b = t.Stats[id.Row].Running
		case ColSucceeded:
			b = t.Stats[id.Row].Succeeded
		case ColFailed:
			b = t.Stats[id.Row].Failed
		case ColTotal:
			b = t.Stats[id.Row].Total
		}

		_, err := b.Get()
		if err == nil {
			l.Bind(binding.IntToString(b))
		}
		updateColumnsWidths(id, o)
	}
	t.Table.UpdateHeader = func(id widget.TableCellID, o fyne.CanvasObject) {
		if id.Row == -1 { // Column Header
			l := o.(*widget.Label)
			l.SetText(StatsTableColumn(id.Col).String())
			updateColumnsWidths(id, o)
		} else { // Row Header
			o.(*widget.Label).SetText(fmt.Sprintf("%d", id.Row+1))
		}
	}
	t.ExtendBaseWidget(t)
	return t
}

func (t *StatsTable) MinSize() fyne.Size {
	minWidth := t.Table.MinSize().Width

	rows, _ := t.Table.Length()
	var totalHeight float32
	padding := theme.Padding()

	if rows > 0 {
		templateCell := t.Table.CreateCell()
		cellHeight := templateCell.MinSize().Height
		totalHeight += (float32(rows) * cellHeight) + (float32(rows-1) * padding)
	}

	if t.Table.ShowHeaderRow {
		var headerHeight float32
		if t.Table.CreateHeader != nil {
			headerTemplate := t.Table.CreateHeader()
			headerHeight = headerTemplate.MinSize().Height
		} else {
			temp := t.Table.CreateCell()
			headerHeight = temp.MinSize().Height
		}

		if totalHeight > 0 {
			totalHeight += headerHeight + padding
		} else {
			totalHeight = headerHeight
		}
	}

	return fyne.NewSize(minWidth, totalHeight)
}

func (t *StatsTable) MouseIn(ev *desktop.MouseEvent)    {}
func (t *StatsTable) MouseDown(e *desktop.MouseEvent)   {}
func (t *StatsTable) MouseMoved(ev *desktop.MouseEvent) {}
func (t *StatsTable) MouseOut()                         {}
func (t *StatsTable) Cursor() desktop.Cursor            { return desktop.DefaultCursor }
func (t *StatsTable) TouchDown(e *mobile.TouchEvent)    {}
func (t *StatsTable) Tapped(e *fyne.PointEvent)         {}
