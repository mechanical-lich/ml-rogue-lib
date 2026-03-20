// Package rltermclient provides a terminal-only ASCII client that works with
// mlge's transport layer. It renders entities carrying AsciiAppearanceComponent
// and PositionComponent as colored characters in a tcell terminal screen.
//
// Unlike rlasciiclient (which requires an Ebiten window), this package has no
// graphical dependency and is suitable for headless environments, SSH sessions,
// or tooling that runs alongside a game server.
//
// Typical setup:
//
//	local := transport.NewLocalTransport()
//	codec := &rlasciiclient.AsciiCodec{}
//
//	tc, err := rltermclient.New(local.Client(), codec)
//	// ...
//	tc.OnInput = func(ev *tcell.EventKey) *transport.Command { ... }
//	tc.Run()
package rltermclient

import (
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlasciiclient"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rltermgui"
	"github.com/mechanical-lich/mlge/transport"
)

// TerminalClient is a self-contained terminal client.
//
// It polls the server transport for snapshots, decodes them into an AsciiWorld,
// renders the viewport each tick, and forwards keyboard input as Commands.
//
// The viewport is sized to the current terminal dimensions automatically.
// Set CameraX/CameraY/CameraZ to control which region is visible.
type TerminalClient struct {
	// World holds the client's decoded entity state.
	World *rlasciiclient.AsciiWorld

	// Camera: tile coordinate of the viewport's top-left corner.
	// Update these each tick (e.g. in OnTick) to follow the player.
	CameraX int
	CameraY int
	CameraZ int

	// Background is the terminal color used for empty cells.
	Background tcell.Color

	// OnInput is called for every key event. Return a *transport.Command to
	// forward it to the server, or nil to ignore. Return a Command with Type
	// set to the special value QuitCommand to stop the client loop.
	OnInput func(ev *tcell.EventKey) *transport.Command

	// OnTick is called once per render tick, after decoding the latest snapshot
	// but before rendering. Use it to update the camera or run client-side logic.
	OnTick func(snap *transport.Snapshot)

	// TickRate is how often the client polls for snapshots and redraws.
	// Defaults to 50ms (20 Hz) if zero.
	TickRate time.Duration

	// GUI is an optional overlay rendered after the world each tick.
	// Views are drawn in registration order; key events are offered to visible
	// views in reverse order (topmost first) before being passed to OnInput.
	// If a view consumes the event (HandleKey returns true), OnInput is skipped.
	GUI *rltermgui.GUI

	t      transport.ClientTransport
	codec  transport.SnapshotCodec
	screen tcell.Screen
}

// QuitCommand is a sentinel Command.Type returned by OnInput to stop Run.
const QuitCommand transport.CommandType = "__quit__"

// New creates a TerminalClient and initialises the tcell screen.
// Call Run to start the event loop. Call Fini (deferred) to restore the terminal
// if you need to clean up without running.
func New(t transport.ClientTransport, codec transport.SnapshotCodec) (*TerminalClient, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite))
	s.Clear()

	return &TerminalClient{
		World:      rlasciiclient.NewAsciiWorld(),
		Background: tcell.ColorBlack,
		t:          t,
		codec:      codec,
		screen:     s,
	}, nil
}

// Fini restores the terminal. Called automatically by Run on exit.
func (c *TerminalClient) Fini() {
	c.screen.Fini()
}

// ScreenSize returns the current terminal dimensions in columns and rows.
// Safe to call from OnTick to compute a centered camera position.
func (c *TerminalClient) ScreenSize() (cols, rows int) {
	return c.screen.Size()
}

// Run starts the client loop and blocks until it exits.
// The loop ends when:
//   - OnInput returns a Command with Type == QuitCommand
//   - A terminal interrupt (Ctrl+C) or resize-to-zero occurs
func (c *TerminalClient) Run() {
	defer c.screen.Fini()

	rate := c.TickRate
	if rate <= 0 {
		rate = 50 * time.Millisecond
	}
	ticker := time.NewTicker(rate)
	defer ticker.Stop()

	// Poll terminal events in a goroutine so they don't block the tick loop.
	eventCh := make(chan tcell.Event, 32)
	go func() {
		for {
			ev := c.screen.PollEvent()
			if ev == nil {
				return
			}
			eventCh <- ev
		}
	}()

	for {
		select {
		case <-ticker.C:
			// Receive and decode the latest snapshot (may be nil).
			var snap *transport.Snapshot
			if raw := c.t.ReceiveSnapshot(); raw != nil {
				snap = raw
				c.codec.Decode(snap, c.World)
			}

			if c.OnTick != nil {
				c.OnTick(snap)
			}

			c.render()
			if c.GUI != nil {
				c.GUI.Draw(c.screen)
			}
			c.screen.Show()

		case ev := <-eventCh:
			switch e := ev.(type) {
			case *tcell.EventKey:
				// GUI views get first crack at input; if consumed, skip OnInput.
				if c.GUI != nil && c.GUI.HandleKey(e) {
					break
				}
				if c.OnInput != nil {
					if cmd := c.OnInput(e); cmd != nil {
						if cmd.Type == QuitCommand {
							return
						}
						c.t.SendCommand(cmd)
					}
				} else {
					// Default: Escape or 'q' quits.
					if e.Key() == tcell.KeyEscape || e.Rune() == 'q' {
						return
					}
				}
			case *tcell.EventResize:
				c.screen.Sync()
			case *tcell.EventInterrupt:
				return
			}
		}
	}
}

// render draws the current world state to the tcell screen.
//
// Two-pass rendering ensures entities always appear on top of tiles:
//  1. Tiles (Blueprint == "tile") populate the background layer.
//  2. Non-tile entities overwrite any tile at the same cell.
//
// This prevents flashing caused by the random iteration order of AsciiWorld.Entities.
func (c *TerminalClient) render() {
	c.screen.Clear()
	w, h := c.screen.Size()

	bg := c.Background
	if bg == tcell.ColorDefault {
		bg = tcell.ColorBlack
	}

	type cell struct {
		char  string
		color tcell.Color
	}
	cells := make(map[[2]int]cell, len(c.World.Entities))

	renderPass := func(tilesOnly bool) {
		for _, e := range c.World.Entities {
			isTile := e.Blueprint == "tile"
			if isTile != tilesOnly {
				continue
			}
			if !e.HasComponent(rlcomponents.Position) || !e.HasComponent(rlcomponents.AsciiAppearance) {
				continue
			}
			pc := e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			if pc.Z != c.CameraZ {
				continue
			}
			col := pc.X - c.CameraX
			row := pc.Y - c.CameraY
			if col < 0 || col >= w || row < 0 || row >= h {
				continue
			}
			ac := e.GetComponent(rlcomponents.AsciiAppearance).(*rlcomponents.AsciiAppearanceComponent)
			cells[[2]int{col, row}] = cell{
				char:  ac.Character,
				color: tcell.NewRGBColor(int32(ac.R), int32(ac.G), int32(ac.B)),
			}
		}
	}

	renderPass(true)  // tiles first (background)
	renderPass(false) // entities second (always on top)

	for key, cell := range cells {
		style := tcell.StyleDefault.Foreground(cell.color).Background(bg)
		c.screen.SetContent(key[0], key[1], []rune(cell.char)[0], nil, style)
	}
}
