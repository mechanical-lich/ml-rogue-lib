package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// StatsComponent holds D&D-style combat statistics.
type StatsComponent struct {
	AC              int
	Str             int
	Dex             int
	Int             int
	Wis             int
	BasicAttackDice string
	BaseDamageType  string   // e.g., "slashing", "piercing", "bludgeoning", "fire"
	Resistances     []string // damage types this entity resists (half damage)
	Weaknesses      []string // damage types this entity is weak to (double damage)
}

func (pc StatsComponent) GetType() ecs.ComponentType {
	return Stats
}
