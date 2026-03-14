package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// InanimateComponent marks an entity as static (trees, rocks, buildings).
// Inanimate entities are stored in a separate list and skipped by turn systems.
type InanimateComponent struct{}

func (pc InanimateComponent) GetType() ecs.ComponentType {
	return Inanimate
}
