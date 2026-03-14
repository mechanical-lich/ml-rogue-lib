package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// PoisonedComponent is a decaying status that deals 1 damage per turn.
type PoisonedComponent struct {
	Duration int
}

func (pc PoisonedComponent) GetType() ecs.ComponentType {
	return Poisoned
}

func (pc *PoisonedComponent) Decay() bool {
	pc.Duration--
	return pc.Duration <= 0
}
