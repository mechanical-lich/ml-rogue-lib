---
layout: default
title: rlcomponents
nav_order: 3
---

# Components

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents`

All ECS component types used by ML Rogue Lib systems. Add these to `*ecs.Entity` instances to opt them into the relevant behaviors.

## Component Type Constants

All constants are of type `ecs.ComponentType` (a `string`).

```go
const (
    Health         = "Health"
    Stats          = "Stats"
    Initiative     = "Initiative"
    MyTurn         = "MyTurn"
    Dead           = "Dead"
    Description    = "Description"
    Solid          = "Solid"
    Inanimate      = "Inanimate"
    NeverSleep     = "NeverSleep"
    Nocturnal      = "Nocturnal"
    WanderAI       = "WanderAI"
    HostileAI      = "HostileAI"
    DefensiveAI    = "DefensiveAI"
    AIMemory       = "AIMemory"
    Alerted        = "Alerted"
    Poisoned       = "Poisoned"
    Poisonous      = "Poisonous"
    Burning        = "Burning"
    Regeneration   = "Regeneration"
    LightSensitive = "LightSensitive"
    Light          = "Light"
    Door           = "Door"
    Food           = "Food"
    Inventory      = "Inventory"
    Item           = "Item"
    Armor          = "Armor"
    Weapon         = "Weapon"
)
```

`Position` and `Direction` are defined in their own files but follow the same convention.

---

## Core Components

### PositionComponent

```go
type PositionComponent struct {
    X, Y, Z int
    Level   int
}
```

Tracks an entity's grid coordinates. `Level` is an optional floor/level index for multi-level dungeons.

**Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetX` | `() int` | Returns the X coordinate |
| `GetY` | `() int` | Returns the Y coordinate |
| `GetZ` | `() int` | Returns the Z coordinate |
| `SetPosition` | `(x, y, z int)` | Updates all three coordinates at once |

---

### HealthComponent

```go
type HealthComponent struct {
    MaxHealth int
    Health    int
    Energy    int
}
```

Tracks hit points and energy. When `Health` drops to ≤ 0, `rlai.HandleDeath` (or `DeadComponent`) marks the entity as dead.

---

### StatsComponent

```go
type StatsComponent struct {
    AC              int
    Str             int
    Dex             int
    Int             int
    Wis             int
    BasicAttackDice string
    BaseDamageType  string
    Resistances     []string
    Weaknesses      []string
}
```

D&D-style combat statistics. `BasicAttackDice` uses MLGE's dice expression format (e.g. `"1d6"`, `"2d8+2"`). `Resistances` and `Weaknesses` hold damage type strings (e.g. `"fire"`, `"slashing"`).

---

### DescriptionComponent

```go
type DescriptionComponent struct {
    Name    string
    Faction string
}
```

Display name and optional faction. Entities sharing the same non-empty `Faction` are considered friendly by `rlcombat.IsFriendly`.

---

### InitiativeComponent

```go
type InitiativeComponent struct {
    DefaultValue  int
    OverrideValue int
    Ticks         int
}
```

Controls turn timing. `Ticks` decrements each frame by `InitiativeSystem.Speed`. When it reaches 0 the entity receives `MyTurn`. Lower `DefaultValue` means more frequent turns. `OverrideValue > 0` replaces `DefaultValue` for the next reset.

---

### MyTurnComponent

A marker component added by `InitiativeSystem` when an entity's turn arrives. Removed by `CleanUpSystem` at the end of each frame. Systems that should act once per turn check for this component.

```go
// GetMyTurn returns a pooled *MyTurnComponent.
func GetMyTurn() *MyTurnComponent
```

---

### DeadComponent

Marker component added when an entity's `Health` reaches zero. `CleanUpSystem` removes entities carrying this component from the level each frame.

---

## AI Components

### WanderAIComponent

Marker component. Entities carrying this component move randomly via `AISystem`.

---

### HostileAIComponent

```go
type HostileAIComponent struct {
    SightRange int
    TargetX    int
    TargetY    int
    Path       []path.Pather
}
```

Causes the entity to pursue and attack the nearest valid target within `SightRange`. `Path` caches the current A* route.

---

### DefensiveAIComponent

```go
type DefensiveAIComponent struct {
    Attacked  bool
    AttackerX int
    AttackerY int
}
```

Set by `rlcombat.TriggerDefenses` when the entity is hit. Causes the entity to turn toward and flee from or retaliate against the attacker position.

---

### AIMemoryComponent

```go
type AIMemoryComponent struct {
    Attacked  bool
    AttackerX int
    AttackerY int
}
```

Persistent attack memory that survives across turns, unlike `DefensiveAIComponent`.

---

### NeverSleepComponent

Marker component. Entities carrying this always receive `MyTurn` regardless of time of day.

---

### NocturnalComponent

Marker component. Entities only receive `MyTurn` when `level.IsNight()` returns `true` (unless also `Alerted`).

---

## Status Effect Components

All status effects implement `DecayingComponent` — a shared interface that decrements a duration counter and returns `true` when the effect has expired.

```go
type DecayingComponent interface {
    Decay() bool
}
```

### AlertedComponent

```go
type AlertedComponent struct { Duration int }
```

Overrides nocturnal/diurnal scheduling: an alerted entity always acts. Decays over `Duration` turns.

---

### PoisonedComponent

```go
type PoisonedComponent struct { Duration int }
```

Deals 1 HP damage per turn via `StatusConditionSystem`. Decays over `Duration` turns.

---

### PoisonousComponent

```go
type PoisonousComponent struct { Duration int }
```

Applied to attackers. When they hit a target, `rlcombat.ApplyStatusEffects` transfers a `PoisonedComponent` to the defender.

---

### BurningComponent

```go
type BurningComponent struct { Duration int }
```

Deals 2 HP damage per turn via `StatusConditionSystem`. Decays over `Duration` turns.

---

### RegenerationComponent

```go
type RegenerationComponent struct { Amount int }
```

Restores `Amount` HP per turn (capped at `MaxHealth`) via `StatusConditionSystem`.

---

### LightSensitiveComponent

Marker component. Handled by `StatusConditionSystem` if a game registers logic for it via `OnStatusEffect`.

---

## Item & Equipment Components

### InventoryComponent

```go
type InventoryComponent struct {
    LeftHand          *ecs.Entity
    RightHand         *ecs.Entity
    Head              *ecs.Entity
    Torso             *ecs.Entity
    Legs              *ecs.Entity
    Feet              *ecs.Entity
    Bag               []*ecs.Entity
    StartingInventory []string
}
```

Holds equipped items in named slots and a list of carried items in `Bag`. `StartingInventory` contains blueprint names to give the entity on spawn.

**Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `AddItem` | `(item *ecs.Entity)` | Appends item to `Bag` |
| `RemoveItem` | `(item *ecs.Entity) bool` | Removes item from `Bag` by pointer |
| `RemoveItemByName` | `(name string) bool` | Removes first item in `Bag` matching `Blueprint` |
| `RemoveAll` | `(name string) bool` | Removes all items matching `DescriptionComponent.Name` |
| `HasItem` | `(name string) bool` | Returns true if `Bag` contains an item with that name |
| `Equip` | `(item *ecs.Entity)` | Equips item to the appropriate slot based on `ItemComponent.Slot` |
| `GetAttackModifier` | `() int` | Sum of attack modifiers from equipped weapons |
| `GetDefenseModifier` | `() int` | Sum of defense modifiers from equipped armor |
| `GetAttackDice` | `() string` | Returns the attack dice of the equipped weapon, or `""` |

---

### ItemComponent

```go
type ItemComponent struct {
    Slot string // e.g. "RightHand", "Head", "Torso"
}
```

Marks an entity as an equippable item. `Slot` determines which `InventoryComponent` slot it occupies when equipped.

---

### WeaponComponent

```go
type WeaponComponent struct {
    AttackDice  string
    AttackBonus int
    DamageType  string
}
```

---

### ArmorComponent

```go
type ArmorComponent struct {
    DefenseBonus int
    Resistances  []string
}
```

---

### FoodComponent

```go
type FoodComponent struct { Amount int }
```

Tracks remaining nutrition. `rlai.Eat` decrements `Amount` by one. `CleanUpSystem` skips removal of dead food entities while `Amount > 0`.

---

## Environment Components

### SolidComponent

Marker component. Entities carrying this block movement.

---

### InanimateComponent

Marker component. Tells the ECS that this entity never acts (static objects). Sets `ecs.InanimateComponentType` automatically via `init()`.

---

### DoorComponent

```go
type DoorComponent struct {
    Open          bool
    Locked        bool
    OpenedSpriteX int
    OpenedSpriteY int
    ClosedSpriteX int
    ClosedSpriteY int
    OwnedBy       string
}
```

`OwnedBy` is a faction or settlement name. Entities belonging to that faction may pass through regardless of `Locked`. `DoorSystem` syncs sprite coordinates to the `Open` state each frame.

---

### LightComponent

Marker component for light-emitting entities. Specific properties depend on game implementation.

---

### DirectionComponent

```go
type DirectionComponent struct { Direction int }
```

Tracks facing direction: `0` = right, `1` = down, `2` = up, `3` = left. Updated by `rlai.Face`.
