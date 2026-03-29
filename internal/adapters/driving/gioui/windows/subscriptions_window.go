package windows

import (
	"fmt"
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/op/paint"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/component"
	"golang.org/x/exp/shiny/materialdesign/icons"

	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/appwidgets"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/adapters/driving/gioui/style"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/common"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/domain"
	"github.com/bluegradienthorizon/singtoolboxgui/internal/core/ports"
	"github.com/google/uuid"
)

type SubItem struct {
	Share    widget.Clickable
	Edit     widget.Clickable
	Delete   widget.Clickable
	Selected widget.Bool
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
	icShare  *widget.Icon

	modal      *appwidgets.Modal
	noteEditor component.TextField
	urlEditor  component.TextField
	saveBtn    widget.Clickable
	cancelBtn  widget.Clickable
	editingID  string
	subs       []domain.Subscription

	// Share modal state
	shareModal    *appwidgets.Modal
	shareQRImage  widget.Image
	shareString   string
	shareCopyBtn  widget.Clickable
	shareCloseBtn widget.Clickable

	// Mass share modal state
	massShareModal         *appwidgets.Modal
	massShareBtn           widget.Clickable
	massShareTab           int // 0 = export, 1 = import
	massShareTabExport     widget.Clickable
	massShareTabImport     widget.Clickable
	massShareQRImage       widget.Image
	massShareString        string
	massShareCopyBtn       widget.Clickable
	massShareCloseBtn      widget.Clickable
	massShareSelectAll     widget.Clickable
	massShareDeselectAll   widget.Clickable
	massShareImportEditor  component.TextField
	massShareImportBtn     widget.Clickable
	massSharePasteBtn      widget.Clickable
	massShareExportList    widget.List
	massShareGenerateQRBtn widget.Clickable
	// QR display modal state
	qrDisplayModal    *appwidgets.Modal
	qrDisplayCloseBtn widget.Clickable
}

func NewSubscriptionsPage(router *Router, c ports.Configuration) *SubscriptionsPage {
	icEdit, _ := widget.NewIcon(icons.EditorModeEdit)
	icDelete, _ := widget.NewIcon(icons.ActionDelete)
	icShare, _ := widget.NewIcon(icons.SocialShare)

	p := &SubscriptionsPage{
		router:         router,
		conf:           c,
		subItems:       make(map[string]*SubItem),
		icEdit:         icEdit,
		icDelete:       icDelete,
		icShare:        icShare,
		modal:          appwidgets.NewModal(),
		shareModal:     appwidgets.NewModal(),
		massShareModal: appwidgets.NewModal(),
		massShareTab:   0,
		qrDisplayModal: appwidgets.NewModal(),
	}
	p.list.List.Axis = layout.Vertical
	p.massShareExportList.Axis = layout.Vertical
	return p
}

func (p *SubscriptionsPage) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.backBtn.Clicked(gtx) {
		p.router.Pop()
	}

	if p.addBtn.Clicked(gtx) {
		p.showEditor(gtx, -1)
	}

	if p.massShareBtn.Clicked(gtx) {
		p.showMassShareModal(gtx)
	}

	renderItem := func(gtx layout.Context, i int) layout.Dimensions {
		s := p.subs[i]
		item, ok := p.subItems[s.ID]
		if !ok {
			item = &SubItem{}
			p.subItems[s.ID] = item
		}

		if item.Share.Clicked(gtx) {
			p.showShareModal(gtx, s)
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

		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Start}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(material.Body1(th, s.Note).Layout),
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								statsText := fmt.Sprintf("Total: %d | Working: %d", len(s.ProfilesURIs), len(s.WorkingProfiles))
								return material.Caption(th, statsText).Layout(gtx)
							}),
						)
					}),
					layout.Rigid(appwidgets.MaterialButton(th, &item.Share, "", p.icShare, "Share").Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Rigid(appwidgets.MaterialButton(th, &item.Edit, "", p.icEdit, "Edit").Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Rigid(appwidgets.MaterialButton(th, &item.Delete, "", p.icDelete, "Delete").Layout),
				)
			}),
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
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Rigid(appwidgets.MaterialButton(th, &p.massShareBtn, "", p.icShare, "Mass Share").Layout),
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
	p.shareModal.Layout(gtx, th)
	p.massShareModal.Layout(gtx, th)
	p.qrDisplayModal.Layout(gtx, th)

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

		gtx.Constraints.Max.X = gtx.Dp(350)
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

// showShareModal displays a modal with QR code for sharing a subscription
func (p *SubscriptionsPage) showShareModal(gtx layout.Context, s domain.Subscription) {
	// Generate share string
	p.shareString = s.Note + "!" + s.URL

	img, err := common.GenerateQRCode(p.shareString)
	if err != nil {
		// If QR generation fails, still show modal with just the text
		p.shareQRImage.Src = paint.ImageOp{}
	} else {
		p.shareQRImage.Src = paint.NewImageOp(img)
	}

	// Show modal
	p.shareModal.Show(gtx, p.shareModalContent)
}

// shareModalContent renders the content of the share modal
func (p *SubscriptionsPage) shareModalContent(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Handle button clicks
	if p.shareCloseBtn.Clicked(gtx) {
		p.shareModal.Dismiss(gtx)
	}

	if p.shareCopyBtn.Clicked(gtx) {
		p.copyShareString(gtx)
	}

	gtx.Constraints.Max.X = gtx.Dp(350)
	return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title
			layout.Rigid(material.H6(th, "Share Subscription").Layout),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// QR Code (centered horizontally)
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if p.shareQRImage.Src == (paint.ImageOp{}) {
					return material.Body1(th, "QR code generation failed").Layout(gtx)
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, p.shareQRImage.Layout)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// Share String (readable text)
			layout.Rigid(material.Body1(th, p.shareString).Layout),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// Buttons
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(1, appwidgets.MaterialButton(th, &p.shareCopyBtn, "Copy", nil, "").Layout),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, appwidgets.MaterialButton(th, &p.shareCloseBtn, "Close", nil, "").Layout),
				)
			}),
		)
	})
}

// copyShareString copies the share string to clipboard
func (p *SubscriptionsPage) copyShareString(gtx layout.Context) {
	go func() {
		gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(p.shareString))})
	}()
}

// showMassShareModal displays a modal for mass import/export of subscriptions
func (p *SubscriptionsPage) showMassShareModal(gtx layout.Context) {
	// Reset selection state when opening modal
	for _, item := range p.subItems {
		item.Selected.Value = false
	}
	p.massShareTab = 0
	p.massShareImportEditor.SetText("")
	p.massShareString = ""
	p.massShareQRImage.Src = paint.ImageOp{}

	p.massShareModal.Show(gtx, p.massShareModalContent)
}

// massShareModalContent renders the content of the mass share modal
func (p *SubscriptionsPage) massShareModalContent(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Handle button clicks
	if p.massShareCloseBtn.Clicked(gtx) {
		p.massShareModal.Dismiss(gtx)
	}

	if p.massShareTabExport.Clicked(gtx) {
		p.massShareTab = 0
	}

	if p.massShareTabImport.Clicked(gtx) {
		p.massShareTab = 1
	}

	if p.massShareSelectAll.Clicked(gtx) {
		for _, item := range p.subItems {
			item.Selected.Value = true
		}
	}

	if p.massShareDeselectAll.Clicked(gtx) {
		for _, item := range p.subItems {
			item.Selected.Value = false
		}
	}

	if p.massShareGenerateQRBtn.Clicked(gtx) {
		p.generateExportQR(gtx)
	}

	if p.massShareCopyBtn.Clicked(gtx) {
		p.copyMassShareString(gtx)
	}

	if p.massSharePasteBtn.Clicked(gtx) {
		p.pasteFromClipboard(gtx)
	}

	if p.massShareImportBtn.Clicked(gtx) {
		p.importSubscriptions(gtx)
	}

	gtx.Constraints.Max.X = gtx.Dp(400)
	gtx.Constraints.Max.Y = gtx.Dp(500)
	return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			// Title
			layout.Rigid(material.H6(th, "Mass Share Subscriptions").Layout),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// Tabs
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &p.massShareTabExport, "Export")
						if p.massShareTab == 0 {
							btn.Background = th.ContrastBg
						} else {
							btn.Background = th.Bg
						}
						return btn.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						btn := material.Button(th, &p.massShareTabImport, "Import")
						if p.massShareTab == 1 {
							btn.Background = th.ContrastBg
						} else {
							btn.Background = th.Bg
						}
						return btn.Layout(gtx)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// Content based on selected tab
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				if p.massShareTab == 0 {
					return p.renderExportTab(gtx, th)
				}
				return p.renderImportTab(gtx, th)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			// Close button
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return appwidgets.MaterialButton(th, &p.massShareCloseBtn, "Close", nil, "").Layout(gtx)
			}),
		)
	})
}

// renderExportTab renders the export tab content
func (p *SubscriptionsPage) renderExportTab(gtx layout.Context, th *material.Theme) layout.Dimensions {
	// Count selected items
	selectedCount := 0
	for _, s := range p.subs {
		if item, ok := p.subItems[s.ID]; ok && item.Selected.Value {
			selectedCount++
		}
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		// Select/Deselect buttons
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, appwidgets.MaterialButton(th, &p.massShareSelectAll, "Select All", nil, "").Layout),
				layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
				layout.Flexed(1, appwidgets.MaterialButton(th, &p.massShareDeselectAll, "Deselect All", nil, "").Layout),
			)
		}),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		// Selected count
		layout.Rigid(material.Body2(th, fmt.Sprintf("Selected: %d subscription(s)", selectedCount)).Layout),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		// Subscription list with checkboxes
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			l := material.List(th, &p.massShareExportList)
			return l.Layout(gtx, len(p.subs), func(gtx layout.Context, i int) layout.Dimensions {
				s := p.subs[i]
				item, ok := p.subItems[s.ID]
				if !ok {
					item = &SubItem{}
					p.subItems[s.ID] = item
				}

				// Handle checkbox update (don't generate QR immediately)
				item.Selected.Update(gtx)

				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.CheckBox(th, &item.Selected, "").Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
							layout.Rigid(material.Body1(th, s.Note).Layout),
							layout.Rigid(material.Caption(th, s.URL).Layout),
						)
					}),
				)
			})
		}),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		// Generate QR Code and Copy buttons (show together when items are selected)
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if selectedCount == 0 {
				return layout.Dimensions{}
			}
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(1, appwidgets.MaterialButton(th, &p.massShareGenerateQRBtn, "Generate QR Code", nil, "").Layout),
				layout.Rigid(layout.Spacer{Width: style.DefaultMargin}.Layout),
				layout.Flexed(1, appwidgets.MaterialButton(th, &p.massShareCopyBtn, "Copy to Clipboard", nil, "").Layout),
			)
		}),
	)
}

// renderImportTab renders the import tab content
func (p *SubscriptionsPage) renderImportTab(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(material.Body1(th, "Paste subscriptions in format: .Note!.URL\\n.Note!.URL\\n...").Layout),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return p.massShareImportEditor.Layout(gtx, th, "Subscriptions to import")
		}),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return appwidgets.MaterialButton(th, &p.massSharePasteBtn, "Paste from Clipboard", nil, "").Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return appwidgets.MaterialButton(th, &p.massShareImportBtn, "Import", nil, "").Layout(gtx)
		}),
	)
}

// generateExportQR generates the export string and QR code based on selected subscriptions
func (p *SubscriptionsPage) generateExportQR(gtx layout.Context) {
	var lines []string
	for _, s := range p.subs {
		item, ok := p.subItems[s.ID]
		if ok && item.Selected.Value {
			lines = append(lines, s.Note+"!"+s.URL)
		}
	}
	p.massShareString = strings.Join(lines, "\n")

	if p.massShareString != "" {
		img, err := common.GenerateQRCode(p.massShareString)
		if err != nil {
			p.massShareQRImage.Src = paint.ImageOp{}
		} else {
			p.massShareQRImage.Src = paint.NewImageOp(img)
		}
	} else {
		p.massShareQRImage.Src = paint.ImageOp{}
		return
	}

	// Show QR code in a new modal
	p.qrDisplayModal.Show(gtx, p.qrDisplayModalContent)
}

// qrDisplayModalContent renders the QR code display modal
func (p *SubscriptionsPage) qrDisplayModalContent(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if p.qrDisplayCloseBtn.Clicked(gtx) {
		p.qrDisplayModal.Dismiss(gtx)
	}

	gtx.Constraints.Max.X = gtx.Dp(300)
	return layout.UniformInset(style.DefaultMargin).Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(material.H6(th, "QR Code").Layout),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				if p.massShareQRImage.Src == (paint.ImageOp{}) {
					return material.Body1(th, "QR code generation failed").Layout(gtx)
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						return layout.Center.Layout(gtx, p.massShareQRImage.Layout)
					}),
				)
			}),
			layout.Rigid(layout.Spacer{Height: style.DefaultMargin}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return appwidgets.MaterialButton(th, &p.qrDisplayCloseBtn, "Close", nil, "").Layout(gtx)
			}),
		)
	})
}

// copyMassShareString copies the mass share string to clipboard
func (p *SubscriptionsPage) copyMassShareString(gtx layout.Context) {
	go func() {
		gtx.Execute(clipboard.WriteCmd{Data: io.NopCloser(strings.NewReader(p.massShareString))})
	}()
}

// pasteFromClipboard pastes content from clipboard to the import editor
func (p *SubscriptionsPage) pasteFromClipboard(gtx layout.Context) {
	// Request clipboard read
	gtx.Execute(clipboard.ReadCmd{Tag: &p.massShareImportEditor})
}

// importSubscriptions parses the import string and adds new subscriptions
func (p *SubscriptionsPage) importSubscriptions(gtx layout.Context) {
	text := p.massShareImportEditor.Text()
	if text == "" {
		return
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse format: .Note!.URL
		parts := strings.SplitN(line, "!", 2)
		if len(parts) != 2 {
			continue
		}

		note := parts[0]
		url := parts[1]

		if note == "" || url == "" {
			continue
		}

		// Add new subscription
		p.subs = append(p.subs, domain.Subscription{
			ID:   uuid.NewString(),
			Note: note,
			URL:  url,
		})
	}

	p.conf.Subscriptions().Set(p.subs)
	p.massShareModal.Dismiss(gtx)
	p.router.WindowInvalidate()
}
