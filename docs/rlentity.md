---
layout: default
title: rlentity
nav_order: 9
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

Checks whether the entity should die and, if so, adds a `DeadComponent` and returns `true`.

**Detection order:**

1. If the entity already has `DeadComponent`, returns `true` immediately.
2. If the entity has `BodyComponent`, checks all parts for a `KillsWhenBroken` or `KillsWhenAmputated` condition.
3. Regardless of whether a `BodyComponent` is present, also checks `HealthComponent.Health <= 0`.

This means an entity can carry both components: lethal body-part damage kills it via step 2, and legacy health reduction kills it via step 3.

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

### FootprintBlockers

```go
func FootprintBlockers(entity *ecs.Entity, level rlworld.LevelInterface, destX, destY, destZ int, buf *[]*ecs.Entity)
```

Appends all solid entities that overlap the entity's footprint at `(destX, destY, destZ)` to `buf`, excluding `entity` itself. Useful for sized entities (those with `SizeComponent`) to enumerate every blocker before deciding whether to attack or open a door.

---

### FindByID

```go
func FindByID(level rlworld.LevelInterface, id string) *ecs.Entity
```

Searches all level entities for the first whose `DescriptionComponent.ID` matches `id`. Returns `nil` if not found or if `id` is empty.

---

### FindByTag

```go
func FindByTag(level rlworld.LevelInterface, tag string) []*ecs.Entity
```

Returns all entities whose `DescriptionComponent.Tags` slice contains `tag`.

---

### CheckInteraction

```go
func CheckInteraction(actor, target *ecs.Entity) bool
```

Fires all `InteractionComponent` triggers on `target` if it has one and has not yet been used (or is marked repeatable). Posts an `InteractionEvent` to MLGE's queued event manager for each trigger, and optionally posts an interaction prompt message. Returns `true` if triggers fired.

---

### CheckPassOver

```go
func CheckPassOver(mover *ecs.Entity, level rlworld.LevelInterface, x, y, z int)
```

Call after a successful `Move`. For each entity at `(x, y, z)` other than the mover:

- Calls `CheckInteraction` (walk-over pressure plates, floor triggers, etc.).
- Posts a random `PassOverDescription` message from the first entity that has one.

---

### CheckExcuseMe

```go
func CheckExcuseMe(bumped *ecs.Entity)
```

Posts a random `ExcuseMeAnnouncement` from `bumped`. Call this after a friendly position swap so the displaced entity can react.

---

### CheckDeathAnnouncement

```go
func CheckDeathAnnouncement(watcher *ecs.Entity, dying *ecs.Entity, level *rlworld.Level)
```

Posts a death message if `watcher` has line-of-sight to `dying` (same Z-level). Uses a random `DeathAnnouncements` string from the dying entity's `DescriptionComponent`, or `"<name> has died."` as a fallback. Call this before the entity is removed from the level.

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
