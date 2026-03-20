---
layout: default
title: rltermgui
nav_order: 14
---

# rltermgui

`github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui`

Provides a lightweight terminal GUI layer for `rltermclient`. It is built on [tcell](https://github.com/gdamore/tcell) and designed around three concepts:

| Type | Role |
|------|------|
| `View` | Interface all GUI elements implement |
| `Pane` | Embeddable rectangular panel with border, title, and show/hide state |
| `GUI` | Manages an ordered list of Views; routes drawing and key input |

---

## View

```go
type View interface {
    Draw(s tcell.Screen)
    HandleKey(ev *tcell.EventKey) bool
    Visible() bool
}
```

Implement `View` to create any GUI element. `Draw` and `HandleKey` are only called when `Visible()` returns `true`. Return `true` from `HandleKey` to consume the event and prevent it from reaching the server.

---

## GUI

```go
type GUI struct {
    Views []View
}
```

### Methods

| Method | Description |
|--------|-------------|
| `(g) Add(v View)` | Appends a view. Views draw in registration order (first = bottom). |
| `(g) Draw(s tcell.Screen)` | Renders all visible views in registration order. |
| `(g) HandleKey(ev *tcell.EventKey) bool` | Routes the event to visible views in **reverse** order (topmost first). Returns `true` if any view consumed it. |
| `(g) AnyVisible() bool` | Reports whether any view is currently visible. Useful for suppressing game-world input while a popup is open. |

### Draw and input order

Views are drawn first-to-last so later views appear on top. Key events are routed last-to-first so the topmost visible view gets first crack at input. If no view consumes the event, `rltermclient` forwards it to `OnInput`.

---

## Pane

`Pane` is an embeddable struct for building Views that occupy a rectangular region of the screen. It provides a border, an optional title, a fill background, and show/hide state.

```go
type Pane struct {
    X, Y, W, H  int
    Title        string
    BorderStyle  tcell.Style  // box outline and title text
    ContentStyle tcell.Style  // interior background fill
}
```

### Functions and methods

| | Description |
|-|-------------|
| `NewPane(x, y, w, h) *Pane` | Creates a Pane with white-on-black default styles. |
| `(p) Show()` | Makes the pane visible. |
| `(p) Hide()` | Makes the pane invisible. |
| `(p) Toggle()` | Flips visibility. |
| `(p) Visible() bool` | Implements `View.Visible`. |
| `(p) DrawPane(s tcell.Screen)` | Fills the content area and draws the border. Call at the top of your `Draw` method. |
| `(p) Inner() (x, y, w, h int)` | Returns the usable content area (inset by 1 on each side). |

### Embedding pattern

```go
type InventoryView struct {
    *rltermgui.Pane
    items []string
}

func (v *InventoryView) Draw(s tcell.Screen) {
    v.DrawPane(s)
    x, y, _, _ := v.Inner()
    for i, item := range v.items {
        rltermgui.DrawText(s, x, y+i, item, v.ContentStyle)
    }
}

func (v *InventoryView) HandleKey(ev *tcell.EventKey) bool {
    if ev.Key() == tcell.KeyEscape {
        v.Hide()
        return true
    }
    return true // consume all input while open
}
```

---

## Drawing helpers

Stateless helper functions for drawing to a `tcell.Screen`:

| Function | Description |
|----------|-------------|
| `DrawText(s, x, y, text, style)` | Writes text horizontally from (x, y). |
| `FillRect(s, x, y, w, h, style)` | Fills a rectangle with spaces. |
| `DrawBox(s, x, y, w, h, title, style)` | Draws a single-line border; centers title in the top edge if non-empty. Does not clear the interior — call `FillRect` first for a solid background. |

---

## Full example

```go
// A simple always-visible HUD showing player HP.
type HUD struct {
    sim *game.SimWorld
}

func (h *HUD) Visible() bool { return true }
func (h *HUD) HandleKey(*tcell.EventKey) bool { return false }
func (h *HUD) Draw(s tcell.Screen) {
    style := tcell.StyleDefault.Foreground(tcell.ColorRed)
    hp := fmt.Sprintf("HP: %d", h.sim.Player.HP)
    rltermgui.DrawText(s, 0, 0, hp, style)
}

// A modal inventory popup.
type InvView struct {
    *rltermgui.Pane
    sim *game.SimWorld
}

func NewInvView(sim *game.SimWorld) *InvView {
    p := rltermgui.NewPane(10, 5, 40, 20)
    p.Title = "Inventory"
    return &InvView{Pane: p, sim: sim}
}

func (v *InvView) Draw(s tcell.Screen) {
    v.DrawPane(s)
    x, y, _, _ := v.Inner()
    for i, item := range v.sim.Player.Items {
        rltermgui.DrawText(s, x, y+i, item.Name, v.ContentStyle)
    }
}

func (v *InvView) HandleKey(ev *tcell.EventKey) bool {
    if ev.Key() == tcell.KeyEscape {
        v.Hide()
    }
    return true // consume all input while open
}

// Wire into TerminalClient.
tc.GUI = &rltermgui.GUI{}
tc.GUI.Add(&HUD{sim: sim})
inv := NewInvView(sim)
tc.GUI.Add(inv)

tc.OnInput = func(ev *tcell.EventKey) *transport.Command {
    if ev.Rune() == 'i' {
        inv.Toggle()
        return nil
    }
    // ... handle movement, etc.
    return nil
}
```
