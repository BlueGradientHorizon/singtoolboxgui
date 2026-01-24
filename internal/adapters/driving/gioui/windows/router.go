package windows

import (
	"gioui.org/app"
	"gioui.org/io/event"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget/material"
)

const (
	RouteMain      = "main"
	RouteHelp      = "help"
	RouteSettings  = "settings"
	RouteSubs      = "subs"
	RouteSubEditor = "sub_editor"
)

type Page interface {
	Layout(gtx layout.Context, th *material.Theme) layout.Dimensions
	OnPushed()
}

type Router struct {
	window  *app.Window
	pages   map[any]Page
	current any
	history []any
	OnExit  func()
}

func NewRouter(window *app.Window) *Router {
	return &Router{
		window: window,
		pages:  make(map[any]Page),
	}
}

func (r *Router) Register(tag any, p Page) {
	r.pages[tag] = p
	if r.current == nil {
		r.current = tag
	}
}

func (r *Router) Push(tag any) {
	if p, ok := r.pages[tag]; ok {
		if r.current != nil {
			r.history = append(r.history, r.current)
		}
		r.current = tag
		p.OnPushed()
	}
}

func (r *Router) Pop() bool {
	if len(r.history) > 0 {
		lastIndex := len(r.history) - 1
		previousTag := r.history[lastIndex]
		r.history = r.history[:lastIndex]
		r.current = previousTag
		return true
	}
	return false
}

func (r *Router) GetPage(tag any) Page {
	return r.pages[tag]
}

func (r *Router) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if len(r.history) > 0 {
		event.Op(gtx.Ops, r)
		for {
			e, ok := gtx.Event(key.Filter{
				Name: key.NameBack,
			})
			if !ok {
				break
			}

			if ev, ok := e.(key.Event); ok {
				if ev.Name == key.NameBack && ev.State == key.Press {
					if r.Pop() {
						gtx.Execute(op.InvalidateCmd{})
					}
				}
			}
		}
	}

	// 2. Render the current page
	if p, ok := r.pages[r.current]; ok {
		return p.Layout(gtx, th)
	}
	return layout.Dimensions{}
}

func (r *Router) WindowInvalidate() {
	r.window.Invalidate()
}
