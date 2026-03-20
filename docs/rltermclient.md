---
layout: default
title: rltermclient
nav_order: 13
---

# rltermclient

`github.com/mechanical-lich/ml-rogue-lib/pkg/rltermclient`

Provides a terminal-only ASCII game client backed by [tcell](https://github.com/gdamore/tcell). It connects to an mlge server via `transport.ClientTransport`, decodes snapshots into an `rlasciiclient.AsciiWorld`, and renders the result as colored characters in the user's terminal. No graphical window or GPU is required.

---

## Overview

`TerminalClient` is the single exported type. It owns:

- a `tcell.Screen` for rendering and input
- an `rlasciiclient.AsciiWorld` for decoded entity state
- an optional `rltermgui.GUI` for HUDs and popups

Set the exported function fields (`OnInput`, `OnTick`) to hook into the client loop, then call `Run()`.

---

## TerminalClient

```go
type TerminalClient struct {
    World      *rlasciiclient.AsciiWorld

    // Camera: tile coordinate of the viewport's top-left corner.
    CameraX int
    CameraY int
    CameraZ int

    Background tcell.Color

    // OnInput is called for every key event not consumed by the GUI.
    // Return a *transport.Command to forward it to the server.
    // Return a Command with Type == QuitCommand to stop the loop.
    OnInput func(ev *tcell.EventKey) *transport.Command

    // OnTick is called once per tick after decoding the latest snapshot
    // but before rendering. Use it to update the camera or run client-side logic.
    OnTick func(snap *transport.Snapshot)

    // TickRate is how often the client polls and redraws. Defaults to 50 ms.
    TickRate time.Duration

    // GUI is an optional overlay rendered after the world each tick.
    // Views draw in registration order; key events route in reverse order.
    // If a view consumes an event, OnInput is skipped for that event.
    GUI *rltermgui.GUI
}
```

### QuitCommand

```go
const QuitCommand transport.CommandType = "__quit__"
```

Return a `*transport.Command{Type: rltermclient.QuitCommand}` from `OnInput` to exit the run loop cleanly.

---

## Functions and methods

| | Description |
|-|-------------|
| `New(t, codec) (*TerminalClient, error)` | Creates a client, initialises the tcell screen. Call `Run` to start. |
| `(c) Run()` | Starts the event/render loop; blocks until the loop exits. Calls `Fini` automatically on exit. |
| `(c) Fini()` | Restores the terminal. Call if you need to clean up without running. |
| `(c) ScreenSize() (cols, rows int)` | Returns current terminal dimensions. Safe to call from `OnTick`. |

---

## Tick loop

Each tick (default 50 ms):

1. Calls `t.ReceiveSnapshot()`. If a snapshot arrived, decodes it into `World` via `codec.Decode`.
2. Calls `OnTick(snap)` — snap is `nil` if no snapshot arrived this tick.
3. Renders the world with a two-pass render (see below).
4. If `GUI` is set, calls `GUI.Draw(screen)` to overlay HUDs and popups.
5. Calls `screen.Show()` to flush the frame.

Key events arrive on a separate goroutine and are processed between ticks:

1. If `GUI` is set, `GUI.HandleKey(ev)` is called first. If it returns `true`, the event is consumed and `OnInput` is skipped.
2. Otherwise `OnInput(ev)` is called. A returned command is forwarded to the server via `t.SendCommand`.
3. If `OnInput` is nil, `Escape` or `q` quits by default.

The loop exits when:
- `OnInput` returns a command with `Type == QuitCommand`
- A terminal interrupt (`Ctrl+C`) is received
- A resize event reports zero dimensions

---

## Two-pass render

To prevent tiles and entities from competing for the same cell (which would cause flickering due to random map-iteration order), rendering is split into two passes:

1. **Tiles** — entities whose `Blueprint == "tile"` are drawn first as the background layer.
2. **Entities** — all other entities are drawn second and always overwrite tiles at the same cell.

This guarantees that entities are always visible on top of tiles, regardless of the order entities appear in `AsciiWorld.Entities`.

---

## Full wiring example

```go
local := transport.NewLocalTransport()
codec := game.NewSPCodec(sim)

server := simulation.NewServer(
    simulation.ServerConfig{TickRate: 20},
    sim,
    func() []*ecs.Entity { return sim.Level.Entities },
    srvT,
    codec,
)
server.SetState(game.NewMainSimState(sim))
go server.Run()

tc, err := rltermclient.New(cliT, codec)
if err != nil {
    log.Fatal(err)
}
defer tc.Fini()

tc.OnTick = func(snap *transport.Snapshot) {
    cols, rows := tc.ScreenSize()
    tc.CameraX = playerX - cols/2
    tc.CameraY = playerY - rows/2
}

tc.OnInput = func(ev *tcell.EventKey) *transport.Command {
    switch ev.Key() {
    case tcell.KeyUp:
        return &transport.Command{Type: "action", Payload: "W"}
    case tcell.KeyEscape:
        return &transport.Command{Type: rltermclient.QuitCommand}
    }
    return nil
}

tc.Run()
server.Stop()
```

---

## Adding a GUI

Assign a `*rltermgui.GUI` and register views before calling `Run`:

```go
tc.GUI = &rltermgui.GUI{}
tc.GUI.Add(myHUD)       // always visible
tc.GUI.Add(myInventory) // shown/hidden on demand
```

See [`rltermgui`](rltermgui.html) for the full GUI API.
