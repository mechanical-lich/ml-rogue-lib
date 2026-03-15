---
layout: default
title: rlsystems
nav_order: 9
---

# Systems

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems`

Turn-based ECS systems. Each system implements MLGE's `ecs.SystemInterface` (except `CleanUpSystem`, which has its own `Update` method) and is registered with a `ecs.SystemManager` or called directly each frame.

All systems expose **extension hook** fields — Go function values you assign at startup to layer game-specific logic on top of the built-in behaviour without subclassing or forking the library.

---

## InitiativeSystem

Ticks entity initiative counters and grants `MyTurn` when the counter reaches zero, respecting nocturnal/diurnal schedules.

```go
type InitiativeSystem struct {
    Speed        int
    OnEntityTurn func(entity *ecs.Entity)
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Speed` | `int` | Amount subtracted from each entity's `InitiativeComponent.Ticks` per frame |
| `OnEntityTurn` | `func(entity *ecs.Entity)` | Called each time an entity receives `MyTurn`. Use for per-turn setup: animation reset, UI updates, etc. |

### Requires

`Initiative`

### Behaviour

1. Decrements `InitiativeComponent.Ticks` by `Speed`.
2. When `Ticks <= 0`, resets to `DefaultValue` (or `OverrideValue` if `> 0`).
3. Checks sleep schedule:
   - `Nocturnal` entities only act when `level.IsNight()`.
   - All other entities only act when it is **not** night.
   - `Alerted` or `NeverSleep` entities bypass the schedule entirely.
4. Adds `GetMyTurn()` to the entity and calls `OnEntityTurn` if set.

---

## AISystem

Handles `WanderAI`, `HostileAI`, and `DefensiveAI` each turn.

```go
type AISystem struct {
    HostileTargetMatch func(self, candidate *ecs.Entity) bool
    GetPath            func(level rlworld.LevelInterface, from, to rlworld.TileInterface, reuse []path.Pather) []path.Pather
    OnWander           func(entity *ecs.Entity)
    OnHostileAttack    func(level rlworld.LevelInterface, attacker, target *ecs.Entity)
}

func NewAISystem() *AISystem
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `HostileTargetMatch` | `func(self, candidate *ecs.Entity) bool` | Returns `true` if `candidate` is a valid attack target for `self`. Default: has `Health`, not `Dead`, different faction. |
| `GetPath` | `func(level, from, to, reuse) []path.Pather` | Pathfinding function used by `HostileAI`. If `nil`, hostile entities fall back to direct delta movement. |
| `OnWander` | `func(entity *ecs.Entity)` | Called after each wander step. |
| `OnHostileAttack` | `func(level, attacker, target *ecs.Entity)` | Called when a hostile entity lands a hit via `rlcombat.Hit`. |

### Requires

`Position`, `MyTurn`

### Behaviour

1. Calls `rlai.HandleDeath` — skips dead entities.
2. **WanderAI**: picks a random cardinal direction, calls `rlai.Move`, then `rlai.Face`. Calls `OnWander` if set.
3. **HostileAI**: searches for the nearest valid target via `GetClosestEntityMatching`. If found, moves toward it (using `GetPath` if provided) and calls `rlcombat.Hit` if adjacent. Calls `OnHostileAttack` on a successful hit.
4. **DefensiveAI**: responds to having been attacked by moving toward the recorded attacker position.

---

## StatusConditionSystem

Ticks decaying status effects and applies their per-turn damage. Also handles `Regeneration`.

```go
type StatusConditionSystem struct {
    ExtraStatuses  map[string]ecs.ComponentType
    OnStatusEffect func(entity *ecs.Entity, effectName string)
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `ExtraStatuses` | `map[string]ecs.ComponentType` | Additional `DecayingComponent` types to tick alongside the built-ins. Key = effect name. |
| `OnStatusEffect` | `func(entity *ecs.Entity, effectName string)` | Called for every active status each turn. Built-in damage runs first; use this for sounds, FX, custom damage, etc. |

### Requires

`Position`, `MyTurn`

### Built-in Effects

| Effect | Damage per turn |
|--------|----------------|
| `"Poisoned"` | −1 HP |
| `"Burning"` | −2 HP |
| `"Alerted"` | 0 (marker only; decays to remove alert state) |

`Regeneration` is handled separately after the status loop: restores `RegenerationComponent.Amount` HP per turn, capped at `MaxHealth`.

### Adding Custom Statuses

```go
statusSystem := &rlsystems.StatusConditionSystem{
    ExtraStatuses: map[string]ecs.ComponentType{
        "Frozen": rlcomponents.Frozen, // your custom component type
    },
    OnStatusEffect: func(entity *ecs.Entity, effectName string) {
        if effectName == "Frozen" {
            // slow the entity, play ice sound, etc.
        }
    },
}
```

---

## DoorSystem

Updates the visual sprite of `Door` entities based on open/closed state.

```go
type DoorSystem struct {
    OnDoorStateChange func(entity *ecs.Entity, open bool)
    AppearanceType    ecs.ComponentType
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `AppearanceType` | `ecs.ComponentType` | Your game's Appearance component type. When set, the system calls `SetSprite` via the `AppearanceUpdater` interface. |
| `OnDoorStateChange` | `func(entity *ecs.Entity, open bool)` | Called every tick with the door's current state. Use for sounds, pathfinding cache invalidation, etc. |

### AppearanceUpdater Interface

```go
type AppearanceUpdater interface {
    SetSprite(x, y int)
}
```

Implement this on your Appearance component to let `DoorSystem` directly set sprite sheet coordinates without depending on a concrete type.

### Requires

`Door` (plus `AppearanceType` if set)

---

## CleanUpSystem

Removes dead entities and strips `MyTurn` from all entities at the end of each frame.

```go
type CleanUpSystem struct {
    OnEntityDead    func(level rlworld.LevelInterface, entity *ecs.Entity)
    OnEntityRemoved func(level rlworld.LevelInterface, entity *ecs.Entity)
    OnEntityCleanup func(level rlworld.LevelInterface, entity *ecs.Entity)
}

func (s *CleanUpSystem) Update(level rlworld.LevelInterface)
```

> **Note:** `CleanUpSystem` does not implement `ecs.SystemInterface`. Call `cleanup.Update(level)` directly, once per frame, before running the other systems.

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `OnEntityDead` | `func(level, entity)` | Called for each dead entity before removal. Use for loot drops, XP awards, death sounds. |
| `OnEntityRemoved` | `func(level, entity)` | Called immediately after `level.RemoveEntity`. Use for secondary cleanup (e.g. custom registries). |
| `OnEntityCleanup` | `func(level, entity)` | Called for every entity each frame, regardless of death state. |

### Behaviour

1. Iterates all entities, stripping `MyTurn` and calling `OnEntityCleanup`.
2. Collects entities carrying `DeadComponent` into a buffer.
3. For each dead entity: calls `OnEntityDead`, skips removal if it is a food entity with `Amount > 0`, then calls `level.RemoveEntity` and `OnEntityRemoved`.
4. Repeats steps 2–3 for static entities.

---

## Typical Frame Order

```go
func (g *Game) Update() error {
    // 1. Strip MyTurn from last frame; remove newly-dead entities.
    g.cleanup.Update(g.level)

    // 2. Run systems for all entities (initiative → AI → status → door).
    g.systemMgr.UpdateSystemsForEntities(g.level, g.level.GetEntities())

    return nil
}
```
