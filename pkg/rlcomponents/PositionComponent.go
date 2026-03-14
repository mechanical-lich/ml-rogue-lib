package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

const Position ecs.ComponentType = "Position"

// PositionComponent .
type PositionComponent struct {
	X, Y, Z int
	Level   int
}

func (pc PositionComponent) GetType() ecs.ComponentType {
	return Position
}

func (pc PositionComponent) GetX() int {
	return pc.X
}
func (pc PositionComponent) GetY() int {
	return pc.Y
}
func (pc PositionComponent) GetZ() int {
	return pc.Z
}

func (pc *PositionComponent) SetPosition(x int, y int, z int) {
	pc.X = x
	pc.Y = y
	pc.Z = z
}
