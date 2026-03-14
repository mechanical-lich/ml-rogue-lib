package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// SolidComponent marks an entity as blocking movement through its tile.
type SolidComponent struct{}

func (pc SolidComponent) GetType() ecs.ComponentType {
	return Solid
}
