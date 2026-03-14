package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

type ItemSlot string

const (
	HandSlot  ItemSlot = "hand"
	HeadSlot  ItemSlot = "head"
	TorsoSlot ItemSlot = "torso"
	LegsSlot  ItemSlot = "legs"
	FeetSlot  ItemSlot = "feet"
	BagSlot   ItemSlot = "bag"
)

// ItemComponent marks an entity as an item that can be picked up and equipped.
type ItemComponent struct {
	Slot   ItemSlot
	Effect string // "heal", "cure", etc.
	Value  int
}

func (ic ItemComponent) GetType() ecs.ComponentType {
	return Item
}
