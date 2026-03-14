package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// InitiativeComponent controls turn-based timing.
// Ticks counts down each frame; when it hits 0 the entity gets MyTurn.
type InitiativeComponent struct {
	DefaultValue  int // ticks between turns (lower = faster)
	OverrideValue int // if > 0, use this instead of DefaultValue
	Ticks         int // current countdown
}

func (pc InitiativeComponent) GetType() ecs.ComponentType {
	return Initiative
}
