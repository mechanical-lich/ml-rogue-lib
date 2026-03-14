---
layout: default
title: rlai
nav_order: 5
---

# AI Navigation Helpers

`github.com/mechanical-lich/mg-rogue-lib/pkg/rlai`

Higher-level AI navigation utilities: target tracking, range checks, and path-following. These are the primitives used by `AISystem` for hostile pursuit and can be called from custom AI behaviours.

For low-level entity manipulation (moving, facing, eating, death detection) see [`rlentity`](rlentity.html).

## Functions

### TrackTarget

```go
func TrackTarget(x, y, z, x2, y2, z2 int) (int, int, int)
```

Returns the unit delta `(dx, dy, dz)` needed to step from `(x,y,z)` toward `(x2,y2,z2)`. Each component is −1, 0, or 1. Useful as a simple direct-movement fallback when pathfinding is not available.

---

### WithinRange

```go
func WithinRange(x, y, z, x2, y2, z2, rangeX, rangeY, rangeZ int) bool
```

Returns `true` if `(x,y,z)` is within the given per-axis range of `(x2,y2,z2)`. Equivalent to a 3D axis-aligned bounding box check.

---

### WithinRangeCardinal

```go
func WithinRangeCardinal(x, y, x2, y2, rangeX, rangeY int) bool
```

Returns `true` if `(x,y)` is reachable from `(x2,y2)` along a pure cardinal line (no diagonals). Either the X coordinates must match and Y must be within `rangeY`, or the Y coordinates must match and X must be within `rangeX`.

---

### MoveTowardsTarget

```go
func MoveTowardsTarget(
    level rlworld.LevelInterface,
    entity *ecs.Entity,
    targetX, targetY, targetZ int,
    getPath func(level rlworld.LevelInterface, from, to rlworld.TileInterface, reuse []path.Pather) []path.Pather,
) bool
```

Moves the entity one step along a cached A* path toward `(targetX, targetY, targetZ)`. Returns `true` if a step was taken.

Requires the entity to have both `PositionComponent` and `AIMemoryComponent`. The path is cached in `AIMemoryComponent.CurrentSteps` and is recomputed when the target changes or the path is exhausted.

**Behaviour:**

1. Compares the stored target against `(targetX, targetY, targetZ)`. Recomputes the path via `getPath` when they differ or fewer than 2 steps remain.
2. Walks forward along the cached path, skipping steps the entity has already reached.
3. If the next step is solid or blocked by a solid entity, clears the cached path and returns `false`.
4. Otherwise calls `rlentity.HandleMovement` to move and face the entity.

---

## Usage Example

```go
import (
    "github.com/mechanical-lich/mg-rogue-lib/pkg/rlai"
    "github.com/mechanical-lich/mg-rogue-lib/pkg/rlentity"
    "github.com/mechanical-lich/mlge/path"
    "github.com/mechanical-lich/mlge/utility"
)

// Hostile entity pursues the player using A* each turn.
func pursuePlayer(level rlworld.LevelInterface, self, player *ecs.Entity, astar *path.AStar) {
    pc := player.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
    moved := rlai.MoveTowardsTarget(level, self, pc.GetX(), pc.GetY(), pc.GetZ(),
        func(lvl rlworld.LevelInterface, from, to rlworld.TileInterface, reuse []path.Pather) []path.Pather {
            result, _, _ := astar.Path(from, to)
            return result
        },
    )
    if !moved {
        // Fallback: direct step.
        selfPC := self.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
        dx, dy, dz := rlai.TrackTarget(selfPC.GetX(), selfPC.GetY(), selfPC.GetZ(),
            pc.GetX(), pc.GetY(), pc.GetZ())
        rlentity.HandleMovement(level, self, dx, dy, dz)
    }
}
```
