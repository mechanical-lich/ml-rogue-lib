package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DeadComponent is a marker added to entities that have died and are pending cleanup.
type DeadComponent struct{}

func (pc DeadComponent) GetType() ecs.ComponentType {
	return Dead
}
