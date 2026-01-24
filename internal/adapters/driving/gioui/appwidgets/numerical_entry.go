package appwidgets

import (
	"strconv"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"gioui.org/x/component"
)

type NumericalEntry struct {
	component.TextField
	Max                int
	AllowLeadingZeroes bool
	def                int
	oldText            string
}

func NewNumericalEntry(def int) *NumericalEntry {
	e := &NumericalEntry{
		def:     def,
		oldText: strconv.Itoa(def),
	}
	e.SetText(e.oldText)
	e.Editor.SingleLine = true
	return e
}

func (e *NumericalEntry) Default() int {
	return e.def
}

func (e *NumericalEntry) Layout(gtx layout.Context, th *material.Theme, hint string) layout.Dimensions {
	return e.TextField.Layout(gtx, th, hint)
}

func (e *NumericalEntry) Update(gtx layout.Context) bool {
	ok := false
	if _, ok = e.Editor.Update(gtx); ok {
		e.validateAndApply()
	}
	return ok
}

func (e *NumericalEntry) validateAndApply() {
	text := e.Text()
	if text == "" {
		e.fallback("")
		return
	}

	val, err := strconv.Atoi(text)
	if err != nil {
		e.fallback(strconv.Itoa(e.def))
		return
	}

	if !e.AllowLeadingZeroes && len(e.oldText) == 0 && text[0] == '0' {
		e.fallback("")
		return
	}

	if !e.AllowLeadingZeroes && len(text) > 0 && text[0] == '0' {
		e.fallback(strconv.Itoa(e.def))
		return
	}

	if e.Max != 0 && val > e.Max {
		e.fallback(strconv.Itoa(e.Max))
		return
	}

	e.oldText = text
}

func (e *NumericalEntry) fallback(s string) {
	e.SetText(s)
	e.Editor.SetCaret(len(s), len(s))
}
