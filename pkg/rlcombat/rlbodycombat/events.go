package rlbodycombat

import "github.com/mechanical-lich/mlge/event"

const CombatEventType event.EventType = "CombatEvent"

// CombatEvent is posted whenever an attack resolves in v2 combat.
// GUIs can listen for this event to display visual effects (floating damage
// numbers, hit animations, etc.) near the world location of the attack.
type CombatEvent struct {
	// World position of the attacker.
	X, Y, Z int

	AttackerName string
	DefenderName string

	// Damage dealt. Zero means the attack missed.
	Damage     int
	DamageType string

	// BodyPart is the name of the body part that was hit.
	// Empty when the attack missed or when the defender has no BodyComponent.
	BodyPart string

	Miss      bool
	Crit      bool
	Broken    bool
	Amputated bool
}

func (e CombatEvent) GetType() event.EventType {
	return CombatEventType
}
