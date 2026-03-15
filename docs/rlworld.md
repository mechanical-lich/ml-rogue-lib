---
layout: default
title: rlworld
nav_order: 4
---

# rlworld

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld`

Provides two interfaces (`LevelInterface`, `TileInterface`) that decouple all library code from any concrete level or tile implementation, plus ready-to-use base types (`Level`, `Tile`, `TileDefinition`) that implement those interfaces with GC-optimized data structures.

---

## Interfaces

### LevelInterface

```go
type LevelInterface interface { ... }
```

Abstracts a 3D tile grid that can hold entities. All systems receive a `LevelInterface` rather than a concrete type.

#### Dimension & Bounds

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetWidth` | `() int` | Width of the grid in tiles |
| `GetHeight` | `() int` | Height of the grid in tiles |
| `GetDepth` | `() int` | Depth (number of Z layers) of the grid |
| `InBounds` | `(x, y, z int) bool` | Returns true if the coordinates are within the grid |

#### Tile Access

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetTileAt` | `(x, y, z int) TileInterface` | Returns the tile at the given coordinates, or `nil` |
| `GetTileIndex` | `(index int) TileInterface` | Returns a tile by flat array index |
| `UpdateTileAt` | `(x, y, z int, tileType string, variant int) TileInterface` | Changes the tile type and returns the updated tile |
| `SetTileType` | `(x, y int, t string) error` | Sets the tile type at (x,y) on Z=0 |
| `AreNeighborsTheSame` | `(t *Tile) (top, bottom, left, right bool)` | Reports whether each cardinal neighbor shares the same Type and Variant as `t` (useful for autotiling) |

#### Time & Lighting

| Method | Signature | Description |
|--------|-----------|-------------|
| `SunIntensity` | `() int` | Current sunlight intensity (0-100) |
| `IsNight` | `() bool` | Returns true when it is currently night — used by `InitiativeSystem` |
| `IsTileExposedToSun` | `(x, y, z int) bool` | Returns true if the tile has direct sunlight (no solid tiles above) |
| `InvalidateSunColumn` | `(x, y int)` | Marks a column as needing a sun-exposure recalculation |
| `NextHour` | `()` | Advances in-game time by one hour |

#### Entity Management

| Method | Signature | Description |
|--------|-----------|-------------|
| `PlaceEntity` | `(x, y, z int, entity *ecs.Entity)` | Moves the entity to (x,y,z), updating its `PositionComponent` and the spatial index |
| `AddEntity` | `(entity *ecs.Entity)` | Registers a new entity on the level (dynamic or static based on `Inanimate` component) |
| `RemoveEntity` | `(entity *ecs.Entity)` | Removes the entity from the level entirely |
| `GetEntityAt` | `(x, y, z int) *ecs.Entity` | Returns the first entity at the coordinates, or `nil` |
| `GetEntitiesAt` | `(x, y, z int, buffer *[]*ecs.Entity)` | Appends all entities at the coordinates to the buffer |
| `GetEntitiesAround` | `(x, y, z, width, height int, buffer *[]*ecs.Entity)` | Appends all entities in the rectangular area to the buffer |
| `GetClosestEntity` | `(x, y, z, width, height int) *ecs.Entity` | Returns the nearest entity within the search rectangle |
| `GetSolidEntityAt` | `(x, y, z int) *ecs.Entity` | Returns the first entity carrying `SolidComponent` at the coordinates |
| `GetClosestEntityMatching` | `(x, y, z, width, height int, exclude *ecs.Entity, match func(*ecs.Entity) bool) *ecs.Entity` | Returns the nearest entity that satisfies `match`, ignoring `exclude` |

#### Entity List Access

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetEntities` | `() []*ecs.Entity` | Returns all dynamic entities on the level |
| `GetStaticEntities` | `() []*ecs.Entity` | Returns all static (inanimate) entities on the level |

---

### TileInterface

```go
type TileInterface interface { ... }
```

Abstracts a single tile in the grid. Tiles must also satisfy `path.Pather` so they can be used directly with MLGE's A* pathfinder.

#### Coordinate & Pathfinding

| Method | Signature | Description |
|--------|-----------|-------------|
| `Coords` | `() (x, y, z int)` | Returns the tile's grid coordinates |
| `PathID` | `() int` | Unique integer ID used by the A* pathfinder |
| `PathNeighborsAppend` | `(neighbors []path.Pather) []path.Pather` | Appends passable neighbors to the slice (zero-allocation) |
| `PathNeighborCost` | `(to path.Pather) float64` | Movement cost to an adjacent tile |
| `PathEstimatedCost` | `(to path.Pather) float64` | Heuristic cost estimate (squared Euclidean distance) |

#### Tile Properties

| Method | Signature | Description |
|--------|-----------|-------------|
| `IsSolid` | `() bool` | True if the tile blocks movement |
| `IsWater` | `() bool` | True if the tile is water |
| `IsAir` | `() bool` | True if the tile is open air |

---

## Base Types

The following concrete types implement the interfaces above. Games can use them directly, embed them in wrapper structs, or ignore them and implement the interfaces from scratch.

### TileDefinition

Describes one category of tile (e.g. "grass", "stone_wall"). Loaded from a JSON file or built programmatically.

```go
type TileDefinition struct {
    Name       string        `json:"name"`
    Solid      bool          `json:"solid"`
    Water      bool          `json:"water"`
    Door       bool          `json:"door"`
    Air        bool          `json:"air"`
    StairsUp   bool          `json:"stairsUp"`
    StairsDown bool          `json:"stairsDown"`
    AutoTile   int           `json:"autoTile"`
    Variants   []TileVariant `json:"variants"`
}

type TileVariant struct {
    Variant int `json:"variant"`
    SpriteX int `json:"spriteX"`
    SpriteY int `json:"spriteY"`
}
```

#### AutoTile Modes

The `AutoTile` field controls how `Level.ResolveVariant` selects the visual variant for a tile:

| Constant | Value | Description |
|----------|-------|-------------|
| `AutoTileNone` | `0` | Default. Uses `tile.Variant` as a direct index into `Variants` |
| `AutoTileWall` | `1` | 2-variant wall. If the bottom neighbor is the same tile type, uses `Variants[0]` (top/connected); otherwise `Variants[1]` (edge/sides) |
| `AutoTileBitmask` | `2` | 4-bit cardinal bitmask. Computes an index from top (bit 0), bottom (bit 1), left (bit 2), right (bit 3) neighbor matching → up to 16 variants |

Games that need additional fields can embed `TileDefinition` in their own struct:

```go
type GameTileDefinition struct {
    rlworld.TileDefinition
    NoBudding bool `json:"no_budding"`
}
```

#### Global Registries

| Variable | Type | Description |
|----------|------|-------------|
| `TileDefinitions` | `[]TileDefinition` | Index-based lookup. `Tile.Type` is an index into this slice. |
| `TileNameToIndex` | `map[string]int` | Maps a tile name to its index |
| `TileIndexToName` | `[]string` | Maps an index back to a tile name |

#### Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `LoadTileDefinitions` | `(path string) error` | Reads a JSON array of definitions from disk and populates the global registries |
| `SetTileDefinitions` | `(defs []TileDefinition)` | Populates the global registries from a slice (for programmatic setup, or syncing from a game's extended definitions) |

---

### Tile

A GC-friendly tile struct with **no pointer fields** — the Go garbage collector skips scanning the entire tile array.

```go
type Tile struct {
    Type       int  // Index into TileDefinitions
    Variant    int  // Visual variant
    LightLevel int  // Cached lighting value
    Idx        int  // Flat index into Level.Data (derives X/Y/Z)
}
```

Coordinates are derived from the flat index via the active level's dimensions, avoiding per-tile X/Y/Z storage. The `Tile` struct implements `TileInterface` and `path.Pather`.

#### Methods

| Method | Description |
|--------|-------------|
| `Coords() (x, y, z int)` | Derives coordinates from `Idx` and the active level dimensions |
| `IsSolid() bool` | Looks up `TileDefinitions[t.Type].Solid` |
| `IsWater() bool` | Looks up `TileDefinitions[t.Type].Water` |
| `IsAir() bool` | Looks up `TileDefinitions[t.Type].Air` |
| `PathID() int` | Returns `Idx` |
| `PathNeighborsAppend(...)` | 6-directional (cardinal + Z up/down). Z neighbors require StairsUp/StairsDown. |
| `PathNeighborCost(...)` | Delegates to `Level.PathCostFunc` if set, otherwise `DefaultPathCost` |
| `PathEstimatedCost(...)` | Squared Euclidean distance heuristic |

#### DefaultPathCost

The built-in path cost function used when `Level.PathCostFunc` is nil:
- Solid or water tiles: cost 5000 (impassable)
- Z-level transitions without stairs: cost 1000
- Otherwise: cost 0

Games should set `Level.PathCostFunc` to add entity-blocking, door logic, faction checks, etc.

---

### Level

A GC-optimized 3D tile container with spatial entity indexing. Implements `LevelInterface`.

```go
type Level struct {
    Data           []Tile
    Entities       []*ecs.Entity
    StaticEntities []*ecs.Entity
    Width, Height, Depth int
    Hour, Day      int
    DirtyColumns   []int
    PathCostFunc   func(from, to *Tile) float64
}
```

#### Construction

```go
level := rlworld.NewLevel(width, height, depth)
```

`NewLevel` allocates the tile array, initializes all tiles to `"air"` in parallel across available CPUs, and sets this level as the **active level** (required for `Tile.Coords()` derivation).

#### Key Design Points

- **Flat 3D array**: Tiles are stored in a single `[]Tile` slice indexed by `x + y*Width + z*Width*Height`. No pointer fields means the GC skips scanning the entire array.
- **Spatial entity index**: An internal `map[int][]*ecs.Entity` maps flat tile indices to entity lists for O(1) lookups.
- **Parallel init**: `InitTiles()` uses `runtime.NumCPU()-1` goroutines to initialize tiles.
- **Active level pattern**: A package-level `activeLevel` pointer is used by `Tile.Coords()` and pathfinding methods. Only one level can be active at a time. Call `level.SetActive()` if you switch between levels.

#### Additional Methods (beyond LevelInterface)

| Method | Signature | Description |
|--------|-----------|-------------|
| `SetActive` | `()` | Makes this level the active level for Tile coordinate derivation |
| `InitTiles` | `()` | Reinitializes all tiles to air (parallel) |
| `GetTilePtr` | `(x, y, z int) *Tile` | Returns a direct `*Tile` pointer (nil if out of bounds) — use this when you need the concrete type |
| `GetTilePtrIndex` | `(idx int) *Tile` | Returns a direct `*Tile` pointer by flat index |
| `ResolveVariant` | `(t *Tile) TileVariant` | Returns the correct `TileVariant` for the tile based on its `AutoTile` mode and neighbors |

#### Embedding the Base Level

Games typically embed `*rlworld.Level` and add rendering or domain-specific fields:

```go
type GameLevel struct {
    *rlworld.Level
    lightOverlay *ebiten.Image
    drawOp       *ebiten.DrawImageOptions
}

func NewGameLevel(w, h, d int) *GameLevel {
    base := rlworld.NewLevel(w, h, d)
    gl := &GameLevel{Level: base}
    // Set custom pathfinding cost for doors, factions, etc.
    base.PathCostFunc = myGamePathCost(gl)
    return gl
}
```

All `LevelInterface` methods are promoted from the embedded base, so `*GameLevel` satisfies `LevelInterface` automatically.

---

## Implementation Notes

- `PlaceEntity` updates the entity's `PositionComponent` and the spatial index, so callers do not need to update position manually.
- `GetSolidEntityAt` is used by `rlai.Move` to detect blocking entities before attempting movement.
- `GetClosestEntityMatching` is used by `AISystem`'s `HostileAI` logic to find attack targets. It searches outward in expanding rings for early exit.
- When implementing `IsAir`: `rlai.Move` treats air tiles as impassable unless the tile directly below is solid, simulating gravity.
- `AreNeighborsTheSame` lives on `LevelInterface` (not `TileInterface`) because it requires access to the tile grid to check neighbors.
