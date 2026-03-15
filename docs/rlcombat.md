---
layout: default
title: rlcombat
nav_order: 6
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
