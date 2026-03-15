---
layout: default
title: rlentity
nav_order: 7
---

# Entity Helpers

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity`

Stateless helper functions for common entity manipulations: movement, facing, interaction, and death detection. These are the low-level primitives used by `AISystem`, `rlai`, and game-specific input handlers.

For higher-level AI navigation (path following, range checks, target tracking) see [`rlai`](rlai.html).

## Functions

### HandleDeath

```go
func HandleDeath(entity *ecs.Entity) bool
```

Checks whether the entity's `HealthComponent.Health` has dropped to zero or below. If so, adds a `DeadComponent` and returns `true`. Does nothing and returns `false` if the entity has no `HealthComponent`.

Call this at the start of an entity's update to bail out early if it just died.

---

### Move

```go
func Move(entity *ecs.Entity, level rlworld.LevelInterface, deltaX, deltaY, deltaZ int) bool
```

Attempts to move the entity by `(deltaX, deltaY, deltaZ)`. Returns `true` if a solid entity was blocking the destination.

**Movement rules:**

- If a `SolidComponent` entity is at the destination, movement is blocked — unless the blocker is a `DoorComponent` the entity is allowed to pass through (checked via `CanPassThroughDoor`).
- If the destination tile is `IsAir`, the entity instead drops to the solid tile directly below (simulates gravity).
- Water tiles block movement entirely.
- Solid tiles block movement entirely.

---

### HandleMovement

```go
func HandleMovement(level rlworld.LevelInterface, entity *ecs.Entity, deltaX, deltaY, deltaZ int)
```

Combines `Move` and `Face` in one call. No-ops if all deltas are zero. Used internally by `rlai.MoveTowardsTarget`.

---

### Face

```go
func Face(entity *ecs.Entity, deltaX, deltaY int)
```

Updates the entity's `DirectionComponent` based on a movement delta. Does nothing if the entity lacks a `DirectionComponent`.

| `deltaX` / `deltaY` | Direction |
|---------------------|-----------|
| `deltaX > 0` | 0 (right) |
| `deltaY > 0` | 1 (down) |
| `deltaY < 0` | 2 (up) |
| `deltaX < 0` | 3 (left) |

---

### Eat

```go
func Eat(entity, foodEntity *ecs.Entity) bool
```

Consumes one unit of food from `foodEntity` by decrementing `FoodComponent.Amount`. Returns `true` on success, `false` if `foodEntity` has no `FoodComponent` or if `entity == foodEntity`.

---

### Swap

```go
func Swap(level rlworld.LevelInterface, entity, entityHit *ecs.Entity)
```

Exchanges the grid positions of two entities. Calls `level.PlaceEntity` for both. Does nothing if they are the same entity.

---

### CanPassThroughDoor

```go
func CanPassThroughDoor(entity *ecs.Entity, door *rlcomponents.DoorComponent) bool
```

Returns `true` if the entity is allowed to move through the given door. An entity may pass if the door is open, or if the door's `OwnedBy` faction matches the entity's `DescriptionComponent.Faction`.

---

### GetName

```go
func GetName(entity *ecs.Entity) string
```

Returns the entity's `DescriptionComponent.Name`, or `"Unknown"` if the component is absent.

---

## Usage Example

```go
import (
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
    "github.com/mechanical-lich/mlge/utility"
)

// Player movement handler.
func movePlayer(player *ecs.Entity, level rlworld.LevelInterface, dx, dy int) {
    if rlentity.HandleDeath(player) {
        return
    }
    hitSolid := rlentity.Move(player, level, dx, dy, 0)
    if hitSolid {
        // bump-attack logic, open door prompt, etc.
    }
    rlentity.Face(player, dx, dy)
}

// Simple wander step.
func wanderStep(entity *ecs.Entity, level rlworld.LevelInterface) {
    dx := utility.GetRandom(-1, 2)
    dy := 0
    if dx == 0 {
        dy = utility.GetRandom(-1, 2)
    }
    rlentity.HandleMovement(level, entity, dx, dy, 0)
}
```
