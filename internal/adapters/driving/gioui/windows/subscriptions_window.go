package windows

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/style"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
	"github.com/google/uuid"
)

type SubItem struct {
	Edit   widget.Clickable
	Delete widget.Clickable
}

type SubscriptionsPage struct {
	router *Router
	conf   ports.Configuration

	list     widget.List
	addBtn   widget.Clickable
	backBtn  widget.Clickable
	subItems map[string]*SubItem
	icEdit   *widget.Icon
	icDelete *widget.Icon

	modal      *appwidgets.Modal
	noteEditor component.TextField
	urlEditor  component.TextField
	saveBtn    widget.Clickable
	cancelBtn  widget.Clickable
	editingID  string
	subs       []domain.Subscription
}

func NewSubscriptionsPage(router *Router, c ports.Configuration) *SubscriptionsPage {
	icEdit, _ := widget.NewIcon(icons.EditorModeEdit)
	icDelete, _ := widget.NewIcon(icons.ActionDelete)

	p := &SubscriptionsPage{
		router:   router,
		conf:     c,
		subItems: make(map[string]*SubItem),
		icEdit:   icEdit,
		icDelete: icDelete,
		modal:    appwidgets.NewModal(),
	}
	p.list.List.Axis = layout.Vertical
	return p
}

func (p *SubscriptionsPage) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.backBtn.Clicked(gtx) {
		p.router.Pop()
	}

	if p.addBtn.Clicked(gtx) {
		p.showEditor(gtx, -1)
	}

	renderItem := func(gtx layout.Context, i int) layout.Dimensions {
		s := p.subs[i]
		item, ok := p.subItems[s.ID]
		if !ok {
			item = &SubItem{}
			p.subItems[s.ID] = item
		}

		if item.Edit.Clicked(gtx) {
			p.showEditor(gtx, i)
		}
		if item.Delete.Clicked(gtx) {
			p.subs = append(p.subs[:i], p.subs[i+1:]...)
			p.conf.Subscriptions().Set(p.subs)
			delete(p.subItems, s.ID)
			p.router.WindowInvalidate()
		}

		return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
			layout.Flexed(1, material.Body1(th, s.Note).Layout),
			layout.Rigid(appwidgets.MaterialButton(th, &item.Edit, "", p.icEdit, "Edit").Layout),
			layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
			layout.Rigid(appwidgets.MaterialButton(th, &item.Delete, "", p.icDelete, "Delete").Layout),
		)
	}

	inset := layout.UniformInset(style.DefaultMargin)
	dims := inset.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Rigid(appwidgets.MaterialButton(th, &p.backBtn, "Back", nil, "").Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Rigid(appwidgets.MaterialButton(th, &p.addBtn, "Add", nil, "").Layout),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				l := material.List(th, &p.list)
				l.AnchorStrategy = material.Overlay
				return l.Layout(gtx, len(p.subs), func(gtx layout.Context, i int) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return renderItem(gtx, i)
						}),
						layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
					)
				})
			}),
		)
	})

	p.modal.Layout(gtx, th)

	return dims
}

func (p *SubscriptionsPage) showEditor(gtx layout.Context, indexClicked int) {
	if indexClicked >= 0 {
		p.noteEditor.SetText(p.subs[indexClicked].Note)
		p.urlEditor.SetText(p.subs[indexClicked].URL)
	} else {
		p.noteEditor.SetText("")
		p.urlEditor.SetText("")
	}

	p.modal.Show(gtx, func(gtx layout.Context, th *material.Theme) layout.Dimensions {
		if p.cancelBtn.Clicked(gtx) {
			p.modal.Dismiss(gtx)
		}

		if p.saveBtn.Clicked(gtx) {
			if indexClicked >= 0 {
				p.subs[indexClicked].Note = p.noteEditor.Text()
				p.subs[indexClicked].URL = p.urlEditor.Text()
			} else {
				p.subs = append(p.subs, domain.Subscription{
					ID:   uuid.NewString(),
					Note: p.noteEditor.Text(),
					URL:  p.urlEditor.Text(),
				})
			}
			p.conf.Subscriptions().Set(p.subs)
			p.modal.Dismiss(gtx)
		}

		gtx.Constraints.Max.X = gtx.Dp(400)
		return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					title := "Add Subscription"
					if p.editingID != "" {
						title = "Edit Subscription"
					}
					return material.H6(th, title).Layout(gtx)
				}),
				layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.noteEditor.Layout(gtx, th, "Note")
				}),
				layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return p.urlEditor.Layout(gtx, th, "URL")
				}),
				layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{}.Layout(gtx,
						layout.Flexed(1, appwidgets.MaterialButton(th, &p.cancelBtn, "Cancel", nil, "").Layout),
						layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
						layout.Flexed(1, appwidgets.MaterialButton(th, &p.saveBtn, "Save", nil, "").Layout),
					)
				}),
			)
		})
	})
}

func (p *SubscriptionsPage) OnPushed() {
	p.subs = p.conf.Subscriptions().Get()
}
