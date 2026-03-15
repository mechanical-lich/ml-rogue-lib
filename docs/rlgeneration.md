---
layout: default
title: rlgeneration
nav_order: 8
---

# Level Generation

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlgeneration`

Procedural generation functions for tile-based levels. Includes primitive shape carvers, random-walk entity and tile clusters, bud-based room growth, recursive subdivision, and Perlin noise terrain generators.

All functions accept a `rlworld.LevelInterface` and are tile-name agnostic — you supply the tile type strings (`"wall"`, `"floor"`, `"water"`, etc.) that match your game's tile registry.

---

## Type Aliases

### EntityFactory

```go
type EntityFactory func(name string, x, y, z int) (*ecs.Entity, error)
```

A function that creates a named entity at the given position using your game's blueprint or factory system. Passed to cluster spawners so the library does not depend on any specific entity creation pattern.

---

## Shape Carvers

### CarveRoom

```go
func CarveRoom(level rlworld.LevelInterface, x, y, z, width, height int, wallType, floorType string, noOverwrite bool)
```

Writes `wallType` on the border and `floorType` in the interior of a rectangle starting at `(x,y,z)`. When `noOverwrite` is `true`, solid tiles are left untouched.

---

### CarveRect

```go
func CarveRect(level rlworld.LevelInterface, x1, y1, x2, y2, z int, wallType, floorType string, noOverwrite bool)
```

Same as `CarveRoom` but defined by two corner coordinates instead of position + size. When `noOverwrite` is `true`, non-solid tiles are skipped.

---

### CarveCircle

```go
func CarveCircle(level rlworld.LevelInterface, z, startX, startY, r int, wallType, floorType string, noOverwrite bool)
```

Carves a circle of radius `r` centred at `(startX, startY)`. The ring at distance `>= r-1` receives `wallType`; the interior receives `floorType`. Tiles outside the circle are untouched. When `noOverwrite` is `true`, solid tiles are skipped.

---

### Room

```go
type Room struct {
    X, Y, GetWidth, GetHeight int
}
```

Represents an axis-aligned rectangular room. Returned by room-building helpers for further processing (e.g. placing doors or populating with entities).

---

## Cluster Spawners

All cluster functions perform a random walk starting at `(x,y,z)`, occasionally stepping ±1 in X or Y each iteration.

### CreateClusterOfTiles

```go
func CreateClusterOfTiles(level rlworld.LevelInterface, x, y, z, size int, tileName string)
```

Places `size` tiles of `tileName` in a random-walk cluster. Out-of-bounds positions are silently skipped.

---

### Create3DClusterOfTiles

```go
func Create3DClusterOfTiles(level rlworld.LevelInterface, x, y, z, size int, tileName string)
```

Same as `CreateClusterOfTiles` but the walk also steps along the Z axis.

---

### CreateClusterOfEntities

```go
func CreateClusterOfEntities(
    level rlworld.LevelInterface,
    x, y, z, size int,
    entityName string,
    factory EntityFactory,
    maxRetries int,
)
```

Spawns `size` entities named `entityName` in a random-walk cluster. Positions that are solid, water, air, or already occupied are retried up to `maxRetries` times before advancing the counter.

---

### CreateClusterOfEntitiesTagged

```go
func CreateClusterOfEntitiesTagged(
    level rlworld.LevelInterface,
    x, y, z, size int,
    entityName string,
    factory EntityFactory,
    maxRetries int,
    onSpawn func(*ecs.Entity),
)
```

Same as `CreateClusterOfEntities` but calls `onSpawn` for each successfully created entity. Use this to attach additional components (e.g. a settlement ownership tag) immediately after spawning.

---

## Room Generation

### BudRooms

```go
func BudRooms(
    level rlworld.LevelInterface,
    z, width, height, numRooms int,
    wallType, floorType string,
    canBud func(rlworld.TileInterface) bool,
)
```

Generates `numRooms` rooms by scanning for solid tiles adjacent to open air and carving a rectangle off them. This produces organic, cave-like dungeons where new rooms "bud" from the walls of existing ones.

`canBud` is a predicate that receives the candidate wall tile. Pass `nil` to allow budding from any solid tile, or use it to block budding on tiles with a `NoBudding` property or similar flag.

---

### RoomIntersects

```go
func RoomIntersects(level rlworld.LevelInterface, x, y, z, width, height int) bool
```

Returns `true` if the candidate rectangle is out of bounds or overlaps any non-air tile. Used internally by `BudRooms` to avoid placing overlapping rooms; you can also call it from custom room placement logic.

---

### BuildRecursiveRoom

```go
func BuildRecursiveRoom(
    level rlworld.LevelInterface,
    x1, y1, x2, y2, z, power int,
    wallType, floorType string,
)
```

Recursively subdivides the rectangle `(x1,y1)–(x2,y2)` into nested rooms connected by corridors, using a port of ToME2's `generate.cc` algorithm. `power` controls subdivision depth — higher values produce more complex layouts.

---

## Terrain Generators

Both terrain functions are multi-threaded using `runtime.NumCPU()-1` workers.

### PerlinConfig

```go
type PerlinConfig struct {
    Alpha int64
    Beta  int64
    N     int32
    Seed  int64 // 0 = use current time
}
```

Controls the Perlin noise parameters. Use `DefaultPerlinConfig()` for the same values as the original fantasy_settlements generators.

```go
func DefaultPerlinConfig() PerlinConfig
```

---

### GenerateOverworldThreaded

```go
type OverworldTileFunc func(x, y, z int, noise float64, startingZ int) string

func GenerateOverworldThreaded(
    level rlworld.LevelInterface,
    startingZ int,
    cfg PerlinConfig,
    tileAt OverworldTileFunc,
)
```

Fills the entire level with terrain using 3D Perlin noise, dispatching work in 128-column chunks across available CPU cores.

`tileAt` is called for every `(x,y,z)` coordinate. Return the tile type string to place, or `""` to leave the tile unchanged.

**Example:**

```go
rlgeneration.GenerateOverworldThreaded(level, startingZ, rlgeneration.DefaultPerlinConfig(),
    func(x, y, z int, noise float64, startingZ int) string {
        if z < startingZ { return "mountain" }
        if z == startingZ {
            if noise < -0.1 { return "beach" }
            return "grass"
        }
        if noise >= 0.2 { return "cliff" }
        return ""
    },
)
```

---

### GenerateIslandThreaded

```go
type IslandSurfaceFunc func(x, y int, noise, normDist float64) string
type IslandFillFunc    func(x, y, z, startingZ int, surfaceType string, noise, normDist float64) string

func GenerateIslandThreaded(
    level rlworld.LevelInterface,
    startingZ int,
    cfg PerlinConfig,
    surfaceAt IslandSurfaceFunc,
    fillAt IslandFillFunc,
)
```

Generates island-shaped terrain using Perlin noise combined with a radial distance falloff that creates ocean around the edges. Runs in two passes:

1. **Surface pass** (single-threaded): calls `surfaceAt` for each `(x,y)` to determine the top-layer tile type. A radial falloff is already applied to the noise value before `surfaceAt` is called.
2. **Fill pass** (multi-threaded): calls `fillAt` for every `(x,y,z)` with the column's surface type, the per-voxel Perlin noise, and the normalized radial distance.

**Example:**

```go
rlgeneration.GenerateIslandThreaded(level, startingZ, rlgeneration.DefaultPerlinConfig(),
    func(x, y int, noise, normDist float64) string {
        if normDist > 0.85 || noise < -0.1 { return "water" }
        if normDist > 0.7                  { return "beach" }
        if noise >= 0.2                    { return "mountain" }
        return "grass"
    },
    func(x, y, z, startingZ int, surfaceType string, noise, normDist float64) string {
        if z < startingZ && surfaceType != "water" { return "dirt" }
        return ""
    },
)
```
