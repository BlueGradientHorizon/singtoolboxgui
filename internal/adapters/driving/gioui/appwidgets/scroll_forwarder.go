package appwidgets

// In Gio, scroll forwarding is handled natively by the layout system.
// We keep this file to maintain structure, but the implementation differs.

type ScrollForwarder struct {
	// Not strictly needed in Gio as nested lists handle scroll bubbles.
}

func NewScrollForwarder() *ScrollForwarder {
	return &ScrollForwarder{}
}
