package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// StatMod describes a single numeric stat change applied by a condition.
// Supported stat names: "ac", "str", "dex", "con", "int", "wis",
// "melee_attack_bonus", "ranged_attack_bonus".
type StatMod struct {
	Stat  string
	Delta int
}

// StatConditionComponent is a general-purpose decaying status effect that
// applies a set of stat modifiers while active and reverts them on expiry.
// It implements both DecayingComponent and ConditionModifier.
//
// Example: hardened condition with +2 AC for 5 turns:
//
//	&StatConditionComponent{
//	    Name:     "Hardened",
//	    Duration: 5,
//	    Mods:     []StatMod{{Stat: "ac", Delta: 2}},
//	}
type StatConditionComponent struct {
	Name     string
	Duration int
	Mods     []StatMod

	applied   bool
	originals map[string]int // stat name → value before modification
}

func (c *StatConditionComponent) GetType() ecs.ComponentType {
	return StatCondition
}

func (c *StatConditionComponent) Decay() bool {
	c.Duration--
	return c.Duration <= 0
}

// ApplyOnce applies all stat modifiers the first time it is called (idempotent).
func (c *StatConditionComponent) ApplyOnce(entity *ecs.Entity) {
	if c.applied {
		return
	}
	c.originals = make(map[string]int, len(c.Mods))
	for _, mod := range c.Mods {
		orig, ok := getStat(entity, mod.Stat)
		if !ok {
			continue
		}
		c.originals[mod.Stat] = orig
		setStat(entity, mod.Stat, orig+mod.Delta)
	}
	c.applied = true
}

// Revert restores all stat values to their pre-condition values.
func (c *StatConditionComponent) Revert(entity *ecs.Entity) {
	if !c.applied {
		return
	}
	for _, mod := range c.Mods {
		if orig, ok := c.originals[mod.Stat]; ok {
			setStat(entity, mod.Stat, orig)
		}
	}
}

// getStat reads a named stat from the entity's StatsComponent.
// Returns (value, true) if the stat exists, (0, false) otherwise.
func getStat(entity *ecs.Entity, stat string) (int, bool) {
	if !entity.HasComponent(Stats) {
		return 0, false
	}
	sc := entity.GetComponent(Stats).(*StatsComponent)
	switch stat {
	case "ac":
		return sc.AC, true
	case "str":
		return sc.Str, true
	case "dex":
		return sc.Dex, true
	case "con":
		return sc.Con, true
	case "int":
		return sc.Int, true
	case "wis":
		return sc.Wis, true
	case "melee_attack_bonus":
		return sc.MeleeAttackBonus, true
	case "ranged_attack_bonus":
		return sc.RangedAttackBonus, true
	}
	return 0, false
}

// setStat writes a named stat on the entity's StatsComponent.
func setStat(entity *ecs.Entity, stat string, val int) {
	if !entity.HasComponent(Stats) {
		return
	}
	sc := entity.GetComponent(Stats).(*StatsComponent)
	switch stat {
	case "ac":
		sc.AC = val
	case "str":
		sc.Str = val
	case "dex":
		sc.Dex = val
	case "con":
		sc.Con = val
	case "int":
		sc.Int = val
	case "wis":
		sc.Wis = val
	case "melee_attack_bonus":
		sc.MeleeAttackBonus = val
	case "ranged_attack_bonus":
		sc.RangedAttackBonus = val
	}
}
