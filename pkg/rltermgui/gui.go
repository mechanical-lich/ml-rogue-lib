// Package rltermgui provides a lightweight terminal GUI layer for rltermclient.
//
// It is built on tcell and designed around three concepts:
//
//  1. View — the interface all GUI elements implement.
//  2. Pane — an embeddable rectangular panel with optional border, title, and
//     show/hide state. Embed *Pane in custom structs to build popups or HUDs.
//  3. GUI — a managed list of Views that draws them in order and routes key
//     events in reverse order so the topmost visible view gets input first.
//
// Integration with rltermclient.TerminalClient:
//
//	tc.GUI = &rltermgui.GUI{}
//	hud := &MyHUD{Pane: rltermgui.NewPane(0, 0, 20, 5)}
//	hud.Show()
//	tc.GUI.Add(hud)
//
// Views are drawn after the world render each tick, overlaying tiles and
// entities. Key events are offered to visible views before being forwarded to
// the server — return true from HandleKey to consume the event.
package rltermgui

import "github.com/gdamore/tcell/v2"

// View is the interface for all GUI elements.
//
// Implement this to create custom HUDs or popups. The simplest approach is to
// embed *Pane and call DrawPane at the top of Draw.
type View interface {
	// Draw renders the view to the screen. Only called when Visible() is true.
	Draw(s tcell.Screen)
	// HandleKey handles a key event. Return true to consume it and prevent it
	// from reaching the server. Only called when Visible() is true.
	HandleKey(ev *tcell.EventKey) bool
	// Visible reports whether this view should be drawn and receive input.
	Visible() bool
}

// GUI manages an ordered list of Views. Assign it to TerminalClient.GUI to
// activate terminal-side HUDs and popups.
type GUI struct {
	Views []View
}

// Add appends a view. Views are drawn in registration order (first = bottom).
func (g *GUI) Add(v View) {
	g.Views = append(g.Views, v)
}

// Draw renders all visible views in registration order.
func (g *GUI) Draw(s tcell.Screen) {
	for _, v := range g.Views {
		if v.Visible() {
			v.Draw(s)
		}
	}
}

// HandleKey routes the event to visible views in reverse order (topmost first).
// Returns true if any view consumed the event.
func (g *GUI) HandleKey(ev *tcell.EventKey) bool {
	for i := len(g.Views) - 1; i >= 0; i-- {
		if v := g.Views[i]; v.Visible() && v.HandleKey(ev) {
			return true
		}
	}
	return false
}

// AnyVisible reports whether any view is currently visible.
// Useful for suppressing game-world input while a popup is open.
func (g *GUI) AnyVisible() bool {
	for _, v := range g.Views {
		if v.Visible() {
			return true
		}
	}
	return false
}

// --- Stateless drawing helpers ---

// DrawText writes text horizontally starting at (x, y) with the given style.
func DrawText(s tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, ch := range text {
		s.SetContent(x+i, y, ch, nil, style)
	}
}

// FillRect fills a rectangle with space characters in the given style.
func FillRect(s tcell.Screen, x, y, w, h int, style tcell.Style) {
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			s.SetContent(x+col, y+row, ' ', nil, style)
		}
	}
}

// DrawBox draws a single-line bordered rectangle using tcell box-drawing runes.
// If title is non-empty it is centered in the top edge. The interior is not
// cleared — call FillRect first for a solid background.
func DrawBox(s tcell.Screen, x, y, w, h int, title string, style tcell.Style) {
	if w < 2 || h < 2 {
		return
	}
	s.SetContent(x, y, tcell.RuneULCorner, nil, style)
	s.SetContent(x+w-1, y, tcell.RuneURCorner, nil, style)
	s.SetContent(x, y+h-1, tcell.RuneLLCorner, nil, style)
	s.SetContent(x+w-1, y+h-1, tcell.RuneLRCorner, nil, style)
	for col := 1; col < w-1; col++ {
		s.SetContent(x+col, y, tcell.RuneHLine, nil, style)
		s.SetContent(x+col, y+h-1, tcell.RuneHLine, nil, style)
	}
	for row := 1; row < h-1; row++ {
		s.SetContent(x, y+row, tcell.RuneVLine, nil, style)
		s.SetContent(x+w-1, y+row, tcell.RuneVLine, nil, style)
	}
	if title != "" {
		max := w - 4
		if max < 1 {
			return
		}
		if len([]rune(title)) > max {
			title = string([]rune(title)[:max])
		}
		col := x + (w-len([]rune(title)))/2
		DrawText(s, col, y, title, style)
	}
}

// --- Pane ---

// Pane is an embeddable rectangular panel with optional border, title, and
// show/hide state.
//
// Embed *Pane in a custom struct to build a View:
//
//	type InventoryView struct {
//	    *rltermgui.Pane
//	    items []string
//	}
//
//	func (v *InventoryView) Draw(s tcell.Screen) {
//	    v.DrawPane(s)
//	    x, y, _, _ := v.Inner()
//	    for i, item := range v.items {
//	        rltermgui.DrawText(s, x, y+i, item, v.ContentStyle)
//	    }
//	}
//
//	func (v *InventoryView) HandleKey(ev *tcell.EventKey) bool {
//	    if ev.Key() == tcell.KeyEscape { v.Hide(); return true }
//	    return false
//	}
type Pane struct {
	X, Y, W, H int
	Title       string
	// BorderStyle is used for the box outline and title text.
	BorderStyle tcell.Style
	// ContentStyle is used to fill the interior background.
	ContentStyle tcell.Style
	visible      bool
}

// NewPane creates a Pane at (x, y) with the given dimensions.
// Default styles are white-on-black.
func NewPane(x, y, w, h int) *Pane {
	s := tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorBlack)
	return &Pane{
		X: x, Y: y, W: w, H: h,
		BorderStyle:  s,
		ContentStyle: s,
	}
}

// Show makes the pane visible.
func (p *Pane) Show() { p.visible = true }

// Hide makes the pane invisible.
func (p *Pane) Hide() { p.visible = false }

// Toggle flips visibility.
func (p *Pane) Toggle() { p.visible = !p.visible }

// Visible implements View.
func (p *Pane) Visible() bool { return p.visible }

// DrawPane fills the content area and draws the border.
// Call this at the start of your custom View's Draw method.
func (p *Pane) DrawPane(s tcell.Screen) {
	FillRect(s, p.X, p.Y, p.W, p.H, p.ContentStyle)
	DrawBox(s, p.X, p.Y, p.W, p.H, p.Title, p.BorderStyle)
}

// Inner returns the (x, y, w, h) of the usable content area inside the border.
func (p *Pane) Inner() (x, y, w, h int) {
	return p.X + 1, p.Y + 1, p.W - 2, p.H - 2
}
