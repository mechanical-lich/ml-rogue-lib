---
layout: default
title: rlasciiclient
nav_order: 12
---

# rlasciiclient

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlasciiclient`

Provides the graphical (Ebiten) ASCII client layer for games built on mlge's server/client transport. It decodes server snapshots into a lightweight entity store (`AsciiWorld`) and renders them as a grid of colored characters using `basicfont.Face7x13`.

---

## Overview

The package has three concerns:

| Type | Role |
|------|------|
| `AsciiWorld` | Client-side entity store; populated by `AsciiCodec.Decode` |
| `AsciiCodec` | Implements `transport.SnapshotCodec`; encodes on the server, decodes on the client |
| `AsciiClientState` | Implements `client.ClientState`; renders the viewport each Ebiten frame |

A game typically embeds `*AsciiClientState` in its own `ClientState` and calls `DrawViewport` from `Draw`.

---

## AsciiWorld

A flat entity store keyed by snapshot ID. Used as the `world` argument passed to `AsciiCodec.Decode`.

```go
type AsciiWorld struct {
    Entities []*ecs.Entity
    // byID — unexported O(1) index
}
```

### Functions

| Function | Description |
|----------|-------------|
| `NewAsciiWorld() *AsciiWorld` | Allocates an empty world |
| `(w) FindOrCreate(id, blueprint string) *ecs.Entity` | Returns the entity with the given snapshot ID, creating it if absent |
| `(w) RemoveNotIn(alive map[string]bool)` | Culls entities whose ID is absent from `alive`; rebuilds `Entities` slice |

---

## AsciiCodec

Implements `transport.SnapshotCodec` for ASCII rendering.

```go
type AsciiCodec struct {
    // EncodeFunc is called by the server each tick.
    // If nil, DefaultEncode is used.
    EncodeFunc func(tick uint64, entities []*ecs.Entity) *transport.Snapshot

    // ExtractPos translates a game-specific position component into (x, y, z).
    // If nil, the codec reads rlcomponents.PositionComponent directly.
    ExtractPos func(comps map[ecs.ComponentType]transport.ComponentData) (x, y, z int, ok bool)
}
```

### Methods

| Method | Description |
|--------|-------------|
| `Encode(tick, entities)` | Delegates to `EncodeFunc` or `DefaultEncode` |
| `Decode(snap, world)` | Applies snapshot to `*AsciiWorld`; creates/updates entities with `AsciiAppearanceComponent` and `PositionComponent` |

### DefaultEncode

Encodes every entity that carries both `rlcomponents.AsciiAppearance` and `rlcomponents.Position`. Uses the entity's memory address (`fmt.Sprintf("%p", e)`) as the snapshot ID — correct for `LocalTransport` (single process). For TCP multiplayer, provide a custom `EncodeFunc` that assigns stable string IDs.

```go
func DefaultEncode(tick uint64, entities []*ecs.Entity) *transport.Snapshot
```

### Custom position types

Games that use their own position component instead of `rlcomponents.PositionComponent` can provide `ExtractPos`:

```go
codec := &rlasciiclient.AsciiCodec{
    ExtractPos: func(comps map[ecs.ComponentType]transport.ComponentData) (x, y, z int, ok bool) {
        raw, found := comps["MyPosition"]
        if !found {
            return 0, 0, 0, false
        }
        p := raw.(*MyPositionComponent)
        return p.TileX, p.TileY, p.Floor, true
    },
}
```

---

## AsciiClientState

Implements `client.ClientState`. Renders a `Cols × Rows` viewport of the `AsciiWorld` as a fixed-size character grid.

```go
type AsciiClientState struct {
    World   *AsciiWorld
    CameraX int        // tile coordinate of the viewport's top-left corner
    CameraY int
    CameraZ int
    Cols    int        // viewport width in character cells
    Rows    int        // viewport height in character cells
    CellW   int        // pixel width per cell (default: DefaultCellW = 7)
    CellH   int        // pixel height per cell (default: DefaultCellH = 13)
    Background color.Color // empty-cell color (default: black)
    Face    text.Face  // glyph face (default: basicfont.Face7x13)
}
```

### Constants

```go
const (
    DefaultCellW = 7   // pixels per glyph column (basicfont.Face7x13)
    DefaultCellH = 13  // pixels per glyph row
)
```

### Functions and methods

| | Description |
|-|-------------|
| `NewAsciiClientState(world, cols, rows)` | Creates a state with `CellW/CellH = 7×13`, black background, and the default face |
| `(s) DrawViewport(screen *ebiten.Image)` | Renders the current viewport; call from a wrapping state's `Draw` |
| `(s) Update(snap)` | No-op — camera and game logic go in the wrapping state |
| `(s) Draw(screen)` | Calls `DrawViewport` |
| `(s) Done()` | Always returns `false`; override by wrapping |

### Wrapping pattern

```go
type MyState struct {
    ascii *rlasciiclient.AsciiClientState
}

func (s *MyState) Update(snap *transport.Snapshot) client.ClientState {
    if snap != nil {
        // camera follows player
        s.ascii.CameraX = playerX - s.ascii.Cols/2
        s.ascii.CameraY = playerY - s.ascii.Rows/2
    }
    return nil
}

func (s *MyState) Draw(screen *ebiten.Image) {
    s.ascii.DrawViewport(screen)
    // draw HUD on top ...
}

func (s *MyState) Done() bool { return false }
```

### Custom font

Replace `Face` with any `text/v2.Face` to use a different font:

```go
face, _ := opentype.Parse(myFontBytes)
state.Face = text.NewGoXFace(face)
state.CellW = 16
state.CellH = 24
```

### Render pipeline

Each `DrawViewport` call:

1. Iterates `World.Entities` and builds a `(col, row) → (char, color)` map for the visible region, culled to `[CameraX … CameraX+Cols) × [CameraY … CameraY+Rows)` on layer `CameraZ`. Only the first entity at each cell is rendered.
2. Fills each cell's background rectangle with `Background`.
3. Draws the glyph using `text.Draw` with `ColorScale` set from the entity's `AsciiAppearanceComponent` RGB.

---

## Full wiring example

```go
local := transport.NewLocalTransport()
codec := &rlasciiclient.AsciiCodec{}

world := rlasciiclient.NewAsciiWorld()
state := rlasciiclient.NewAsciiClientState(world, 80, 40)

c := client.NewClient(
    local.Client(), codec, state, world,
    func() []*ecs.Entity { return world.Entities },
    client.ClientConfig{ScreenWidth: 560, ScreenHeight: 520},
)
c.Run()
```
