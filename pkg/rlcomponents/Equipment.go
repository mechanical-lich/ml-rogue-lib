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
	// Display is used as the sprite lookup key; falls back to blueprint name if empty.
	Display string
	// TwoHanded weapons occupy both hand slots; a second weapon cannot be dual-wielded with one.
	TwoHanded bool
}

func (pc WeaponComponent) GetType() ecs.ComponentType {
	return Weapon
}

// ArmorComponent holds defense stats for an armor entity in an inventory slot.
type ArmorComponent struct {
	DefenseBonus  int
	Resistances   []string
	StoppingPower int // SP value: damage absorbed per hit (effective damage = max(0, Pen - SP))
	// Tags are used for equipment appearance row scoring.
	Tags []string
}

func (pc ArmorComponent) GetType() ecs.ComponentType {
	return Armor
}
