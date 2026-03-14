package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// WeaponComponent holds combat stats for a weapon entity in an inventory slot.
type WeaponComponent struct {
	AttackBonus        int
	AttackDice         string
	DamageType         string
	Range              int
	Ranged             bool
	ProjectileX        int
	ProjectileY        int
	ProjectileResource string
}

func (pc WeaponComponent) GetType() ecs.ComponentType {
	return Weapon
}

// ArmorComponent holds defense stats for an armor entity in an inventory slot.
type ArmorComponent struct {
	DefenseBonus int
	Resistances  []string
}

func (pc ArmorComponent) GetType() ecs.ComponentType {
	return Armor
}
