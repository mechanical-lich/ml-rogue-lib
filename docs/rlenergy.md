---
layout: default
title: rlenergy
nav_order: 15
---

# Energy & Turn Management

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy`

Stateless helper functions for running a CDDA-style energy/action-point turn system. Entities accumulate energy each tick and spend it on actions with variable costs. Leftover energy carries over, enabling multi-action turns for fast entities.

This package operates on `*ecs.Entity` values that carry an `EnergyComponent` (from `rlcomponents`). It does not own any state itself — all mutations happen through the ECS.

## How It Works

1. **Each tick**, call `AdvanceEnergy` to add `Speed` to every entity's `Energy`.
2. Entities with `Energy > 0` receive `MyTurn` and can act.
3. When an entity acts, the game sets `LastActionCost` via `SetActionCost`, then adds `TurnTaken`.
4. Call `ResolveTurn` (or a cleanup system that calls it) to deduct the cost and strip turn markers.
5. If an entity still has `Energy > 0` after cost deduction, `RegrantTurns` gives it another turn — no time passes.
6. When no one has leftover energy, go back to step 1.

```
AdvanceEnergy ──► entity acts ──► SetActionCost ──► ResolveTurn
      ▲                                                  │
      │              RegrantTurns (Energy > 0?) ◄────────┘
      │                     │ no
      └─────────────────────┘
```

## Functions

### AdvanceEnergy

```go
func AdvanceEnergy(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool)
```

Adds `Speed` to `Energy` for every entity with an `EnergyComponent`. Entities that can act (`Energy > 0`) and don't already have `MyTurn` receive it. Returns whether the player and/or any entity got a turn.

---

### RegrantTurns

```go
func RegrantTurns(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool)
```

Re-grants `MyTurn` to entities that still have enough energy after their previous action. No energy is ticked — this only checks leftover energy. Use this for multi-action turns.

---

### ResolveTurn

```go
func ResolveTurn(entity *ecs.Entity) bool
```

End-of-turn bookkeeping for a single entity. If the entity has both `MyTurn` and `TurnTaken`, it calls `SpendTurn` on the `EnergyComponent` and removes both markers. Returns `true` if a turn was resolved.

---

### CanAct

```go
func CanAct(entity *ecs.Entity) bool
```

Returns `true` if the entity has an `EnergyComponent` with `Energy > 0`.

---

### GrantTurn

```go
func GrantTurn(entity *ecs.Entity) bool
```

Adds `MyTurn` to the entity if it can act and doesn't already have one. Returns `true` if a turn was granted.

---

### SetActionCost

```go
func SetActionCost(entity *ecs.Entity, cost int)
```

Records the energy cost of the action an entity just took. The cost is consumed by `ResolveTurn` (or `SpendTurn` directly).

---

### MoveCost

```go
func MoveCost(tile *rlworld.Tile, baseCost int) int
```

Returns the energy cost for moving onto the given tile. Multiplies `baseCost` by the tile's `MovementCost` (from `TileDefinition`). A `MovementCost` of `0` is treated as `1`.

## Example: Game Loop Integration

```go
// Player phase — process input immediately, deduct cost same tick.
sim.UpdatePlayer()
rlenergy.ResolveTurn(player)

if rlenergy.CanAct(player) {
    // Player has leftover energy — act again (multi-action).
    player.AddComponent(rlcomponents.GetMyTurn())
} else {
    // Advance time so NPCs accumulate energy.
    playerGotTurn, _ := rlenergy.AdvanceEnergy(entities, player)
    if !playerGotTurn {
        // Enter NPC phase.
    }
}

// NPC phase — check for multi-action, then advance time.
playerGotTurn, anyGotTurn := rlenergy.RegrantTurns(entities, player)
if !anyGotTurn {
    turnCount++
    playerGotTurn, anyGotTurn = rlenergy.AdvanceEnergy(entities, player)
}
```

## Setting Action Costs

Define game-specific cost constants and call `SetActionCost` after each action:

```go
const (
    CostMove   = 100
    CostAttack = 100
    CostQuick  = 50
)

// After a successful move:
destTile := level.GetTilePtr(x, y, z)
rlenergy.SetActionCost(entity, rlenergy.MoveCost(destTile, CostMove))

// After an attack:
rlenergy.SetActionCost(entity, CostAttack)
```

## Blueprint Configuration

`EnergyComponent` is configured per entity in blueprint JSON:

```json
{
    "EnergyComponent": {
        "Speed": 25,
        "Energy": 100
    }
}
```

| Field | Description |
|-------|-------------|
| `Speed` | Energy gained per tick. Higher = faster entity. With a base action cost of 100, a speed of 25 means 4 ticks per action; 50 means 2 ticks. |
| `Energy` | Starting energy. Set to a positive value (e.g. 100) for entities that should act on the first tick. Leave at 0 for entities that must accumulate first. |
