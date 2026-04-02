---
layout: default
title: rlcomponents
nav_order: 7
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
    Energy         = "Energy"
    Initiative     = "Initiative"
    MyTurn         = "MyTurn"
    TurnTaken      = "TurnTaken"
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
    Haste          = "Haste"
    Slowed         = "Slowed"
    DamageCondition = "DamageCondition"
    StatCondition  = "StatCondition"
    Regeneration   = "Regeneration"
    LightSensitive = "LightSensitive"
    Light          = "Light"
    Door           = "Door"
    Food           = "Food"
    Inventory      = "Inventory"
    Item           = "Item"
    Armor          = "Armor"
    Weapon         = "Weapon"
    Body           = "Body"
    BodyInventory  = "BodyInventory"
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

### EnergyComponent

```go
type EnergyComponent struct {
    Speed          int
    Energy         int
    LastActionCost int
}
```

Tick-up action point system. Each tick, `Energy` increases by `Speed`. The entity can act whenever `Energy > 0`. Actions set `LastActionCost`, which is deducted by `SpendTurn`. Leftover energy carries over, enabling multi-action turns. See [`rlenergy`](rlenergy.html) for the turn management helpers that drive this component.

**Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `CanAct` | `() bool` | Returns `true` if `Energy > 0` |
| `SpendTurn` | `() int` | Deducts `LastActionCost` from `Energy`, resets `LastActionCost`, returns the cost |

---

### TurnTakenComponent

Marker component added by game systems when an entity has consumed its turn. `ResolveTurn` (in `rlenergy`) checks for both `MyTurn` and `TurnTaken` before deducting cost and stripping markers.

```go
func GetTurnTaken() *TurnTakenComponent
```

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

A marker component added by `InitiativeSystem` or `rlenergy.AdvanceEnergy` when an entity's turn arrives. Removed by cleanup logic at the end of each frame. Systems that should act once per turn check for this component.

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

Effects that modify entity stats while active also implement `ConditionModifier`. The `StatusConditionSystem` calls `ApplyOnce` on the first tick and `Revert` when the component is removed.

```go
type ConditionModifier interface {
    ApplyOnce(entity *ecs.Entity)
    Revert(entity *ecs.Entity)
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

### HasteComponent

```go
type HasteComponent struct { Duration int }
```

Doubles the entity's `EnergyComponent.Speed` while active. Implements `ConditionModifier` — speed is doubled on the first tick and restored exactly when the component expires. Decays over `Duration` turns.

---

### SlowedComponent

```go
type SlowedComponent struct { Duration int }
```

Halves the entity's `EnergyComponent.Speed` while active (minimum 1). Implements `ConditionModifier` — speed is halved on the first tick and restored exactly when the component expires. Decays over `Duration` turns.

---

### DamageConditionComponent

```go
type DamageConditionComponent struct {
    Name       string
    Duration   int
    DamageDice string
    DamageType string
}
```

A general-purpose damage-over-time effect. Each turn `StatusConditionSystem` calls `Roll()` and routes the result through the standard damage path (random body part if `BodyComponent` is present, otherwise `HealthComponent`).

`DamageDice` uses MLGE's dice expression format (e.g. `"1d6"`, `"2d4+1"`, `"3"`). `DamageType` is informational (e.g. `"poison"`, `"fire"`).

`Name` is displayed in status UIs.

**Method:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `Roll` | `() int` | Evaluates `DamageDice` and returns the result (minimum 1). Returns 1 on a parse error. |

**Example:**

```go
// 1d4 venom damage for 6 turns
entity.AddComponent(&rlcomponents.DamageConditionComponent{
    Name:       "Venom",
    Duration:   6,
    DamageDice: "1d4",
    DamageType: "poison",
})
```

---

### StatConditionComponent

```go
type StatMod struct {
    Stat  string
    Delta int
}

type StatConditionComponent struct {
    Name     string
    Duration int
    Mods     []StatMod
}
```

A general-purpose temporary stat modifier. Implements `ConditionModifier` — all `Mods` are applied on the first tick and reverted exactly when the component expires.

Supported `Stat` values: `"ac"`, `"str"`, `"dex"`, `"con"`, `"int"`, `"wis"`, `"melee_attack_bonus"`, `"ranged_attack_bonus"`.

`Name` is displayed in status UIs.

**Examples:**

```go
// +2 AC for 5 turns ("Hardened")
entity.AddComponent(&rlcomponents.StatConditionComponent{
    Name:     "Hardened",
    Duration: 5,
    Mods:     []rlcomponents.StatMod{{Stat: "ac", Delta: 2}},
})

// +3 STR, −1 DEX for 8 turns ("Hormonal Surge")
entity.AddComponent(&rlcomponents.StatConditionComponent{
    Name:     "Hormonal Surge",
    Duration: 8,
    Mods: []rlcomponents.StatMod{
        {Stat: "str", Delta: 3},
        {Stat: "dex", Delta: -1},
    },
})
```

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

### BodyComponent

```go
type BodyPart struct {
    Name                string
    Description         string
    AttachedTo          []string   // parts this part connects to (informational)
    HP                  int
    MaxHP               int
    Broken              bool
    Amputated           bool
    KillsWhenBroken     bool
    KillsWhenAmputated  bool
    CompatibleItemSlots []ItemSlot
}

type BodyComponent struct {
    Parts map[string]BodyPart
}
```

Replaces (or supplements) `HealthComponent` with a per-part damage model. Each named part tracks its own HP, break state, and amputation state.

**Methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `AddPart` | `(part BodyPart)` | Registers a body part by `part.Name`; initialises `Parts` if nil |

**Part lifecycle:**

- A part becomes **Broken** when `HP` drops to zero or below.
- A part becomes **Amputated** when a single hit deals `damage >= MaxHP * 2`.
- If `KillsWhenBroken` or `KillsWhenAmputated` is `true`, a `DeadComponent` is added to the entity when that condition is met.

**Death detection with `BodyComponent`:**

`rlentity.HandleDeath` checks vital part conditions first (broken/amputated with `KillsWhen*`). If none are met it falls through to the `HealthComponent` check, so both systems can coexist on the same entity.

**Status effect damage with `BodyComponent`:**

`StatusConditionSystem` routes poison/burning damage to a random non-amputated body part. If all parts are amputated it falls back to `HealthComponent`.

---

### BodyInventoryComponent

```go
type BodyInventoryComponent struct {
    Equipped          map[string]*ecs.Entity // keyed by BodyPart.Name
    Bag               []*ecs.Entity
    StartingInventory []string
}
```

An inventory whose equipment slots map directly to body-part names. Use alongside `BodyComponent`. A body part accepts an item when the item's `ItemComponent.Slot` is listed in `BodyPart.CompatibleItemSlots`.

**Bag methods** (mirror `InventoryComponent`):

| Method | Signature | Description |
|--------|-----------|-------------|
| `AddItem` | `(item *ecs.Entity)` | Appends to `Bag` |
| `RemoveItem` | `(item *ecs.Entity) bool` | Removes by pointer |
| `RemoveItemByName` | `(name string) bool` | Removes first match by `Blueprint` name |
| `RemoveAll` | `(name string) bool` | Removes all matching `DescriptionComponent.Name` |
| `HasItem` | `(name string) bool` | Returns true if `Bag` contains a matching item |

**Equip methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `EquipToBodyPart` | `(item *ecs.Entity, partName string)` | Places item in the named part slot; bumps existing item to `Bag` |
| `AutoEquip` | `(item *ecs.Entity, bc *BodyComponent) bool` | Finds the first compatible, non-amputated slot; prefers empty slots |
| `Unequip` | `(partName string) *ecs.Entity` | Returns equipped item to `Bag`; returns nil if empty |
| `UnequipAll` | `()` | Returns all equipped items to `Bag` |
| `HandleAmputation` | `(partName string) *ecs.Entity` | Unequips the amputated part's item |
| `EquipBest` | `(slot ItemSlot, bc *BodyComponent)` | Re-equips the highest-value bag item for the given slot |
| `EquipAllBest` | `(bc *BodyComponent)` | Calls `EquipBest` for every slot accepted by any body part |

**Combat stat methods:**

| Method | Signature | Description |
|--------|-----------|-------------|
| `GetAttackModifier` | `() int` | Sum of `WeaponComponent.AttackBonus` from all equipped weapons |
| `GetAttackDice` | `() string` | Combined dice string for all equipped weapons |
| `GetDefenseModifier` | `() int` | Sum of `ArmorComponent.DefenseBonus` from all equipped armor |
| `GetDamageType` | `() string` | Damage type of the first equipped weapon, or `""` |
| `GetResistances` | `() []string` | All resistance types from all equipped armor |

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
