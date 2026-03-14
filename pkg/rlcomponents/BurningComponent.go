package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// BurningComponent is a decaying status that deals 2 fire damage per turn.
type BurningComponent struct {
	Duration int
}

func (pc BurningComponent) GetType() ecs.ComponentType {
	return Burning
}

func (pc *BurningComponent) Decay() bool {
	pc.Duration--
	return pc.Duration <= 0
}
