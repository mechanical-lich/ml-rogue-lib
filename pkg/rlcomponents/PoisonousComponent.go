package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// PoisonousComponent marks an entity as able to inflict Poisoned on hit.
type PoisonousComponent struct {
	Duration int // duration of the poison applied to the target
}

func (pc PoisonousComponent) GetType() ecs.ComponentType {
	return Poisonous
}
