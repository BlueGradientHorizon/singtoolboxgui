package appwidgets

import (
	"strconv"

	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

func NewPreferenceNumericalEntry(p domain.ConfigurationValue[int]) *NumericalEntry {
	e := NewNumericalEntry(p.Default)
	e.SetText(strconv.Itoa(p.Get()))
	e.OnChanged = func(v string) {
		if v == "" {
			p.Set(p.Default)
			return
		}
		i, _ := strconv.Atoi(v)
		p.Set(i)
	}
	return e
}

func NewPreferenceCheck(label string, p domain.ConfigurationValue[bool]) *widget.Check {
	c := widget.NewCheck(label, func(b bool) { p.Set(b) })
	c.Checked = p.Get()
	return c
}
