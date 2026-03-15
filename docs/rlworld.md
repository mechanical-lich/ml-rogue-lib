---
layout: default
title: rlworld
nav_order: 4
---

# World Interfaces

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld`

Two interfaces that decouple all library code from any concrete level or tile implementation. Implement these on your game's own types to make them compatible with every system and helper in this library.

## LevelInterface

```go
type LevelInterface interface { ... }
```

Abstracts a 3D tile grid that can hold entities. All systems receive a `LevelInterface` rather than a concrete type.

### Dimension Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetWidth` | `() int` | Width of the grid in tiles |
| `GetHeight` | `() int` | Height of the grid in tiles |
| `GetDepth` | `() int` | Depth (number of Z layers) of the grid |
| `InBounds` | `(x, y, z int) bool` | Returns true if the coordinates are within the grid |

### Tile Access

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetTileAt` | `(x, y, z int) TileInterface` | Returns the tile at the given coordinates, or `nil` |
| `GetTileIndex` | `(index int) TileInterface` | Returns a tile by flat array index |
| `UpdateTileAt` | `(x, y, z int, tileType string, variant int) TileInterface` | Changes the tile type and returns the updated tile |
| `SetTileType` | `(x, y int, t string) error` | Sets the tile type at (x,y) on the default Z layer |

### Time & Lighting

| Method | Signature | Description |
|--------|-----------|-------------|
| `SunIntensity` | `() int` | Current sunlight intensity (game-defined scale) |
| `IsNight` | `() bool` | Returns true when it is currently night — used by `InitiativeSystem` |
| `IsTileExposedToSun` | `(x, y, z int) bool` | Returns true if the tile has direct sunlight |
| `InvalidateSunColumn` | `(x, y int)` | Marks a column as needing a sun-exposure recalculation |
| `NextHour` | `()` | Advances in-game time by one hour |

### Entity Management

| Method | Signature | Description |
|--------|-----------|-------------|
| `PlaceEntity` | `(x, y, z int, entity *ecs.Entity)` | Moves the entity to (x,y,z), updating its `PositionComponent` |
| `AddEntity` | `(entity *ecs.Entity)` | Adds a new entity to the level without placing it at a specific tile |
| `RemoveEntity` | `(entity *ecs.Entity)` | Removes the entity from the level entirely |
| `GetEntityAt` | `(x, y, z int) *ecs.Entity` | Returns the first entity at the coordinates, or `nil` |
| `GetEntitiesAt` | `(x, y, z int, buffer *[]*ecs.Entity)` | Appends all entities at the coordinates to the buffer |
| `GetEntitiesAround` | `(x, y, z, width, height int, buffer *[]*ecs.Entity)` | Appends all entities in the rectangular area to the buffer |
| `GetClosestEntity` | `(x, y, z, width, height int) *ecs.Entity` | Returns the nearest entity within the search rectangle |
| `GetSolidEntityAt` | `(x, y, z int) *ecs.Entity` | Returns the first entity carrying `SolidComponent` at the coordinates |
| `GetClosestEntityMatching` | `(x, y, z, width, height int, exclude *ecs.Entity, match func(*ecs.Entity) bool) *ecs.Entity` | Returns the nearest entity that satisfies `match`, ignoring `exclude` |

### Entity List Access

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetEntities` | `() []*ecs.Entity` | Returns all dynamic entities on the level |
| `GetStaticEntities` | `() []*ecs.Entity` | Returns all static (inanimate) entities on the level |

---

## TileInterface

```go
type TileInterface interface { ... }
```

Abstracts a single tile in the grid. Tiles must also satisfy `path.Pather` so they can be used directly with MLGE's A* pathfinder.

### Coordinate & Pathfinding

| Method | Signature | Description |
|--------|-----------|-------------|
| `Coords` | `() (x, y, z int)` | Returns the tile's grid coordinates |
| `PathID` | `() int` | Unique integer ID used by the A* pathfinder |
| `PathNeighborsAppend` | `(neighbors []path.Pather) []path.Pather` | Appends passable neighbors to the slice |
| `PathNeighborCost` | `(to path.Pather) float64` | Movement cost to an adjacent tile |
| `PathEstimatedCost` | `(to path.Pather) float64` | Heuristic cost estimate (e.g. Manhattan distance) |
| `AreNeighborsTheSame` | `() (top, bottom, left, right bool)` | Reports whether each cardinal neighbor shares the same tile type (useful for autotiling) |

### Tile Properties

| Method | Signature | Description |
|--------|-----------|-------------|
| `IsSolid` | `() bool` | True if the tile blocks movement |
| `IsWater` | `() bool` | True if the tile is water (entities cannot enter without special logic) |
| `IsAir` | `() bool` | True if the tile is open air (entities fall to the solid tile below) |

---

## Implementation Notes

- `PlaceEntity` is expected to also update the entity's `PositionComponent`, so callers do not need to update position manually.
- `GetSolidEntityAt` is used by `rlai.Move` to detect blocking entities before attempting movement.
- `GetClosestEntityMatching` is used by `AISystem`'s `HostileAI` logic to find attack targets.
- When implementing `IsAir`: `rlai.Move` treats air tiles as impassable unless the tile directly below is solid, simulating gravity.
