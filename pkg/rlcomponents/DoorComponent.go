package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DoorComponent represents a door that can be opened/closed and optionally owned by a faction.
type DoorComponent struct {
	Open          bool
	Locked        bool
	OpenedSpriteX int
	OpenedSpriteY int
	ClosedSpriteX int
	ClosedSpriteY int
	OwnedBy       string // faction or settlement name that may pass freely
}

func (d *DoorComponent) GetType() ecs.ComponentType {
	return Door
}
