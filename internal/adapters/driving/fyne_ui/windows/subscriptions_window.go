package windows

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
	"github.com/google/uuid"
)

func editSubscriptionPopup(parent fyne.Window, c ports.Configuration, editIdx *int, onDone func()) {
	noteEntry := widget.NewEntry()
	urlEntry := widget.NewEntry()
	title := "Add Subscription"

	subs := c.Subscriptions().Get()
	if editIdx != nil {
		title = "Edit Subscription"
		noteEntry.SetText(subs[*editIdx].Note)
		urlEntry.SetText(subs[*editIdx].URL)
	}

	formItems := []*widget.FormItem{
		{Text: "Note", Widget: noteEntry},
		{Text: "URL", Widget: urlEntry},
	}
	onSubmit := func(submit bool) {
		if !submit {
			return
		}
		if editIdx != nil {
			subs[*editIdx].Note = noteEntry.Text
			subs[*editIdx].URL = urlEntry.Text
		} else {
			subs = append(subs, domain.Subscription{
				ID:   uuid.NewString(),
				Note: noteEntry.Text,
				URL:  urlEntry.Text,
			})
		}
		c.Subscriptions().Set(subs)
		if onDone != nil {
			onDone()
		}
	}

	d := dialog.NewForm(title, "Save", "Cancel", formItems, onSubmit, parent)
	d.Show()
}

func ShowSubscriptions(a fyne.App, c ports.Configuration) {
	w := a.NewWindow("Subscriptions")
	list := container.NewVBox()

	var renderList func()

	renderList = func() {
		list.Objects = nil

		subs := c.Subscriptions().Get()
		for i := range subs {
			sub := &subs[i]
			note := widget.NewLabel(sub.Note)
			edit := widget.NewButtonWithIcon("", theme.DocumentCreateIcon(), func() {
				editSubscriptionPopup(w, c, &i, renderList)
			})
			del := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				dialog.ShowConfirm("Confirm", "Remove this subscription?", func(b bool) {
					if b {
						subs = append(subs[:i], subs[i+1:]...)
						c.Subscriptions().Set(subs)
						renderList()
					}
				}, w)
			})

			list.Add(container.NewHBox(note, layout.NewSpacer(), edit, del))
		}
		list.Refresh()
	}

	renderList()

	addBtn := widget.NewButton("Add", func() {
		editSubscriptionPopup(w, c, nil, renderList)
	})

	w.SetContent(container.NewBorder(nil, addBtn, nil, nil, container.NewScroll(list)))
	w.Resize(fyne.NewSize(400, 400))
	w.Show()
}
