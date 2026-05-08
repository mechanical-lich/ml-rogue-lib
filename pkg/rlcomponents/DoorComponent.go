package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DoorComponent represents a door that can be opened/closed and optionally owned by a faction.
type DoorComponent struct {
	Open            bool
	Locked          bool
	KeyId           string   // if set, the ID of the key that can unlock this door
	OpenedSpriteX   int
	OpenedSpriteY   int
	ClosedSpriteX   int
	ClosedSpriteY   int
	OwnedBy         string   // primary faction/settlement that may pass freely
	AllowedFactions []string // additional factions allowed to pass
	AutoOpened      bool     // true when opened by the proximity system (not manually)
}

func (d *DoorComponent) GetType() ecs.ComponentType {
	return Door
}
