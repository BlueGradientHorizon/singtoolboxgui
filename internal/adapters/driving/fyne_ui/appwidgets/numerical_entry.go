package appwidgets

import (
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type NumericalEntry struct {
	widget.Entry
	Max                int
	AllowLeadingZeroes bool
	def                int
	oldText            string
	oldCol             int
	origOnChanged      func(string)
}

func (e *NumericalEntry) Default() int {
	return e.def
}

func NewNumericalEntry(def int) *NumericalEntry {
	entry := &NumericalEntry{}
	entry.def = def
	entry.ExtendBaseWidget(entry)
	return entry
}

func (e *NumericalEntry) TypedRune(r rune) {
	if r < '0' || r > '9' {
		return
	}
	e.applyAndValidate(func() { e.Entry.TypedRune(r) })
}

func (e *NumericalEntry) TypedKey(k *fyne.KeyEvent) {
	if k.Name == fyne.KeyBackspace || k.Name == fyne.KeyDelete {
		e.applyAndValidate(func() {
			e.Entry.TypedKey(k)
		})
		return
	}

	e.Entry.TypedKey(k)
}

func (e *NumericalEntry) TypedShortcut(shortcut fyne.Shortcut) {
	if _, ok := shortcut.(*fyne.ShortcutPaste); ok {
		e.applyAndValidate(func() { e.Entry.TypedShortcut(shortcut) })
		return
	}
	e.Entry.TypedShortcut(shortcut)
}

func (e *NumericalEntry) applyAndValidate(fn func()) {
	e.oldText = e.Text
	e.oldCol = e.CursorColumn
	e.origOnChanged = e.OnChanged

	e.OnChanged = nil

	fn()

	if e.Text == "" {
		e.fallback("")
		return
	}

	val, err := strconv.Atoi(e.Text)
	if err != nil {
		e.fallback(strconv.Itoa(e.def))
		return
	}

	if !e.AllowLeadingZeroes && len(e.oldText) == 0 && e.Text[0] == '0' {
		e.fallback("")
		return
	}

	if !e.AllowLeadingZeroes && len(e.Text) > 0 && e.Text[0] == '0' {
		e.fallback(strconv.Itoa(e.def))
		return
	}

	if e.Max != 0 && val > e.Max {
		e.fallback(strconv.Itoa(e.Max))
		return
	}

	newText := e.Text
	e.OnChanged = e.origOnChanged
	if e.OnChanged != nil {
		e.OnChanged(newText)
	}
}

func (e *NumericalEntry) fallback(s string) {
	e.OnChanged = e.origOnChanged
	e.Text = s
	e.CursorColumn = len(s)
	e.OnChanged(s)
	e.Refresh()
}
