package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// RegenerationComponent heals the entity by Amount HP each turn.
type RegenerationComponent struct {
	Amount int
}

func (pc RegenerationComponent) GetType() ecs.ComponentType {
	return Regeneration
}
