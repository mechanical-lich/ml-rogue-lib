package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

const Direction ecs.ComponentType = "Direction"

// DirectionComponent .
type DirectionComponent struct {
	Direction int
}

func (pc DirectionComponent) GetType() ecs.ComponentType {
	return Direction
}
