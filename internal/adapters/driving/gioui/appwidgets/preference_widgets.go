package appwidgets

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
)

type PreferenceNumericalEntryStyle struct {
	*NumericalEntry
	pref  domain.ConfigurationValue[int]
	th    *material.Theme
	label string
	hint  string
}

func PreferenceNumericalEntry(th *material.Theme, label, hint string, p domain.ConfigurationValue[int]) *PreferenceNumericalEntryStyle {
	e := NewNumericalEntry(p.Default)
	e.SetText(strconv.Itoa(p.Get()))

	return &PreferenceNumericalEntryStyle{
		NumericalEntry: e,
		pref:           p,
		th:             th,
		label:          label,
		hint:           hint,
	}
}

func (e PreferenceNumericalEntryStyle) Layout(gtx layout.Context) layout.Dimensions {
	return DetailRow{}.Layout(gtx,
		material.Body1(e.th, e.label).Layout,
		func(gtx layout.Context) layout.Dimensions {
			maxWidth := gtx.Dp(unit.Dp(150))
			if gtx.Constraints.Max.X > maxWidth {
				gtx.Constraints.Max.X = maxWidth
			}
			gtx.Constraints.Min.X = 0
			return e.NumericalEntry.Layout(gtx, e.th, e.hint)
		},
	)
}

func (e *PreferenceNumericalEntryStyle) Update(gtx layout.Context) {
	if !e.NumericalEntry.Update(gtx) {
		return
	}
	current := e.Text()
	if current == "" {
		e.pref.Set(e.pref.Default)
		return
	}
	i, _ := strconv.Atoi(current)
	e.pref.Set(i)
}

func (e *PreferenceNumericalEntryStyle) SetCurrentPrefValue() {
	e.SetText(strconv.Itoa(e.pref.Get()))
}

type PreferenceCheckStyle struct {
	Switch *widget.Bool
	th     *material.Theme
	label  string
	pref   domain.ConfigurationValue[bool]
}

func PreferenceCheck(th *material.Theme, swtch *widget.Bool, label string, p domain.ConfigurationValue[bool]) *PreferenceCheckStyle {
	swtch.Value = p.Get()
	return &PreferenceCheckStyle{
		th:     th,
		Switch: swtch,
		label:  label,
		pref:   p,
	}
}

func (c *PreferenceCheckStyle) Layout(gtx layout.Context) layout.Dimensions {
	return DetailRow{}.Layout(gtx,
		material.Body1(c.th, c.label).Layout,
		func(gtx layout.Context) layout.Dimensions {
			return material.Switch(c.th, c.Switch, c.label).Layout(gtx)
		},
	)
}

func (c *PreferenceCheckStyle) Update(gtx layout.Context) {
	if c.Switch.Update(gtx) {
		c.pref.Set(c.Switch.Value)
	}
}
