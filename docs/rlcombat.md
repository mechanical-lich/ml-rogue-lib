---
layout: default
title: rlcombat
nav_order: 8
---

# Combat

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat`

A D&D-style melee combat pipeline covering to-hit rolls, damage calculation, resistances, weaknesses, and status effect transfer. All functions are stateless — wire them into your game systems or player action handlers.

## Constants

```go
const DefaultDamageType = "bludgeoning"
```

Used when an entity's `StatsComponent.BaseDamageType` is empty.

---

## Functions

### Hit

```go
func Hit(level rlworld.LevelInterface, entity, entityHit *ecs.Entity, swap bool)
```

Performs a full melee attack from `entity` against `entityHit`.

**Pipeline:**

1. If the two entities are friendly (`IsFriendly` returns `true`) and `swap` is `true`, their positions are exchanged and the function returns.
2. Requires `StatsComponent` on both entities and `HealthComponent` on the defender — returns early otherwise.
3. Rolls `1d20 + Dex modifier + inventory attack bonus` vs `defender AC + inventory defense bonus`.
4. On a hit: calls `InflictDamage`, then `ApplyStatusEffects`.
5. On a miss: posts a tagged "missed" message to MLGE's message log.
6. Always calls `TriggerDefenses` on `entityHit`.

---

### InflictDamage

```go
func InflictDamage(attacker, defender *ecs.Entity)
```

Rolls damage and applies it to the defender's `HealthComponent`.

**Pipeline:**

1. Calls `GetAttackDice` to obtain the dice string, damage type, and Strength modifier.
2. Rolls the dice using MLGE's `dice.ParseDiceRequest`.
3. Halves damage if the defender has a matching resistance (via `StatsComponent.Resistances` or equipped armor).
4. Doubles damage if the defender has a matching weakness.
5. Enforces a minimum of 1 damage.
6. Posts a tagged "combat" message with attacker name, damage amount, and damage type.

---

### GetAttackDice

```go
func GetAttackDice(entity *ecs.Entity) (dice string, damageType string, modifier int)
```

Returns the attack dice expression, damage type, and Strength-based modifier for an entity. If the entity has an `InventoryComponent` with a weapon equipped in `RightHand`, the weapon's dice and damage type override the entity's base stats. Inventory attack bonuses are added to the modifier.

---

### IsFriendly

```go
func IsFriendly(attacker, defender *ecs.Entity) bool
```

Returns `true` if both entities share the same non-empty `DescriptionComponent.Faction`. Used by `Hit` to swap friendly entities instead of attacking them.

---

### TriggerDefenses

```go
func TriggerDefenses(defender *ecs.Entity, attackerX, attackerY int)
```

Notifies the defender that it was attacked:

- Sets `DefensiveAIComponent.Attacked = true` and records attacker coordinates.
- Sets `AIMemoryComponent.Attacked = true` and records attacker coordinates.
- Adds an `AlertedComponent` (duration 120) if not already present.

---

### ApplyStatusEffects

```go
func ApplyStatusEffects(attacker, defender *ecs.Entity)
```

Transfers status effects from attacker to defender on a successful hit. Currently: if the attacker has `PoisonousComponent`, a `PoisonedComponent` is added to the defender (if not already poisoned).

---

### GetModifier

```go
func GetModifier(stat int) int
```

Returns the D&D-style ability modifier: `(stat - 10) / 2`.

---

### IsInArrowPath

```go
func IsInArrowPath(aX, aY, tX, tY, maxRange int) bool
```

Returns `true` if the target at `(tX,tY)` is reachable from `(aX,aY)` via a straight or diagonal line of at most `maxRange` tiles. Useful for determining whether a ranged attack is geometrically possible without a full ray-cast.

---

### HasResistance

```go
func HasResistance(defender *ecs.Entity, damageType string) bool
```

Returns `true` if the entity resists `damageType` — either via `StatsComponent.Resistances` or via `ArmorComponent.Resistances` on any equipped item in the legacy `InventoryComponent` slots.

---

### HasWeakness

```go
func HasWeakness(defender *ecs.Entity, damageType string) bool
```

Returns `true` if the entity is weak to `damageType` via `StatsComponent.Weaknesses`.

---

## Resistances and Weaknesses

Resistances and weaknesses are stored as `[]string` on `StatsComponent`. Equipped armor pieces may also contribute resistances via `ArmorComponent.Resistances`. The damage type is matched case-sensitively.

```go
// Common damage type strings (not constants — define your own as needed).
"bludgeoning"
"slashing"
"piercing"
"fire"
"cold"
"poison"
```

---

## Usage Example

```go
import "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"

// Player bumps into an enemy entity.
func onPlayerBump(level rlworld.LevelInterface, player, target *ecs.Entity) {
    rlcombat.Hit(level, player, target, false)

    // Check if target died.
    if target.HasComponent(rlcomponents.Dead) {
        // award XP, play sound, etc.
    }
}
```

---

# Combat v2

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/v2`

An extended combat pipeline that routes damage to individual body parts when the defender has a `BodyComponent`. Falls back to the legacy `HealthComponent` path for entities without body parts (or when all parts are amputated). Import this package in place of `rlcombat` when using the body system.

## Hit

```go
func Hit(level rlworld.LevelInterface, entity, entityHit *ecs.Entity, swap bool) bool
```

Performs a full melee attack. Returns `true` if the attack was executed.

**Pipeline:**

1. If both entities share a faction and `swap` is `true`, their positions are exchanged; returns `false`.
2. Returns `false` if either entity lacks `StatsComponent`, or if the defender has neither `BodyComponent` nor `HealthComponent`.
3. Rolls `1d20 + Dex modifier + weapon attack bonus` vs `defender AC + armor defense bonus`.
4. **Natural 20** is always a critical hit (doubles damage).
5. On a **hit**:
   - If the defender has a `BodyComponent`, a random non-amputated part is chosen via `randomBodyPart`. Damage is applied via `applyBodyPartDamage`, which sets `Broken` and `Amputated` flags and checks `KillsWhen*`.
   - If all parts are amputated (or no `BodyComponent`), damage is applied to `HealthComponent` instead.
   - If a lethal condition is met, `HealthComponent.Health` is set to `0` and a `DeadComponent` is added.
   - A `CombatEvent` is queued with full hit details.
   - `ApplyStatusEffects` is called.
6. On a **miss**: a "missed" message is posted and a miss `CombatEvent` is queued.
7. Always calls `TriggerDefenses` on the defender.

**Damage routing summary:**

| Defender state | Damage target |
|----------------|--------------|
| Has `BodyComponent`, parts available | Random non-amputated body part |
| Has `BodyComponent`, all parts amputated | `HealthComponent` (fallback) |
| No `BodyComponent`, has `HealthComponent` | `HealthComponent` |
| Neither | Attack is invalid; returns `false` |

---

## CombatEvent

```go
type CombatEvent struct {
    X, Y, Z      int    // world position of the attacker
    AttackerName string
    DefenderName string
    Damage       int    // 0 on a miss
    DamageType   string
    BodyPart     string // empty on miss or health-only hit
    Miss         bool
    Crit         bool
    Broken       bool
    Amputated    bool
}

const CombatEventType event.EventType = "CombatEvent"
```

Posted to MLGE's queued event system on every attack resolution. Register a listener to drive visual effects, floating damage numbers, or sound cues:

```go
import (
    v2 "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/v2"
    "github.com/mechanical-lich/mlge/event"
)

type fxHandler struct{}

func (h *fxHandler) HandleEvent(e event.EventData) error {
    ce, ok := e.(v2.CombatEvent)
    if !ok || ce.Miss {
        return nil
    }
    spawnHitParticle(ce.X, ce.Y, ce.Z, ce.DamageType, ce.Crit)
    return nil
}

// At startup:
event.GetQueuedInstance().RegisterListener(&fxHandler{}, v2.CombatEventType)
```

---

## Usage Example

```go
import (
    v2 "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat/v2"
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
)

// Entity with a body takes damage; vitals check kills it.
func onPlayerBump(level rlworld.LevelInterface, player, target *ecs.Entity) {
    v2.Hit(level, player, target, false)

    if target.HasComponent(rlcomponents.Dead) {
        // drop loot, award XP, play death sound
    }
}
```
