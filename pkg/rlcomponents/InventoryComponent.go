package rlcomponents

import (
	"math/rand/v2"

	"github.com/mechanical-lich/mlge/ecs"
)

// InventoryComponent holds equipped items and a bag for carried items.
type InventoryComponent struct {
	LeftHand          *ecs.Entity
	RightHand         *ecs.Entity
	Head              *ecs.Entity
	Torso             *ecs.Entity
	Legs              *ecs.Entity
	Feet              *ecs.Entity
	Bag               []*ecs.Entity
	StartingInventory []string // blueprint names to give on spawn
}

func (pc InventoryComponent) GetType() ecs.ComponentType {
	return Inventory
}

func (ic *InventoryComponent) AddItem(item *ecs.Entity) {
	ic.Bag = append(ic.Bag, item)
}

func (ic *InventoryComponent) RemoveItem(item *ecs.Entity) bool {
	for i, v := range ic.Bag {
		if v == item {
			ic.Bag = append(ic.Bag[:i], ic.Bag[i+1:]...)
			return true
		}
	}
	return false
}

func (ic *InventoryComponent) RemoveItemByName(name string) bool {
	for i := len(ic.Bag) - 1; i >= 0; i-- {
		if ic.Bag[i].Blueprint == name {
			ic.Bag = append(ic.Bag[:i], ic.Bag[i+1:]...)
			return true
		}
	}
	return false
}

func (ic *InventoryComponent) RemoveAll(name string) bool {
	removed := false
	for i := len(ic.Bag) - 1; i >= 0; i-- {
		item := ic.Bag[i]
		if item.HasComponents(Item, Description) {
			dc := item.GetComponent(Description).(*DescriptionComponent)
			if dc.Name == name {
				ic.Bag = append(ic.Bag[:i], ic.Bag[i+1:]...)
				removed = true
			}
		}
	}
	return removed
}

func (ic *InventoryComponent) HasItem(name string) bool {
	for _, item := range ic.Bag {
		if item.HasComponents(Item, Description) {
			dc := item.GetComponent(Description).(*DescriptionComponent)
			if dc.Name == name {
				return true
			}
		}
	}
	return false
}

func (ic *InventoryComponent) Equip(item *ecs.Entity) {
	if !item.HasComponent(Item) {
		return
	}
	itemC := item.GetComponent(Item).(*ItemComponent)
	switch itemC.Slot {
	case HandSlot:
		if ic.RightHand == nil {
			ic.RightHand = item
		} else if ic.LeftHand == nil {
			ic.LeftHand = item
		} else {
			ic.AddItem(ic.RightHand)
			ic.RightHand = item
		}
		ic.RemoveItem(item)
	case HeadSlot:
		if ic.Head != nil {
			ic.AddItem(ic.Head)
		}
		ic.Head = item
		ic.RemoveItem(item)
	case TorsoSlot:
		if ic.Torso != nil {
			ic.AddItem(ic.Torso)
		}
		ic.Torso = item
		ic.RemoveItem(item)
	case LegsSlot:
		if ic.Legs != nil {
			ic.AddItem(ic.Legs)
		}
		ic.Legs = item
		ic.RemoveItem(item)
	case FeetSlot:
		if ic.Feet != nil {
			ic.AddItem(ic.Feet)
		}
		ic.Feet = item
		ic.RemoveItem(item)
	}
}

func (ic *InventoryComponent) Unequip(slot ItemSlot) *ecs.Entity {
	var item *ecs.Entity
	switch slot {
	case HandSlot:
		item = ic.RightHand
		ic.RightHand = nil
	case HeadSlot:
		item = ic.Head
		ic.Head = nil
	case TorsoSlot:
		item = ic.Torso
		ic.Torso = nil
	case LegsSlot:
		item = ic.Legs
		ic.Legs = nil
	case FeetSlot:
		item = ic.Feet
		ic.Feet = nil
	}
	if item != nil {
		ic.AddItem(item)
	}
	return item
}

func (ic *InventoryComponent) UnequipAll() {
	ic.Unequip(HandSlot)
	ic.Unequip(HeadSlot)
	ic.Unequip(TorsoSlot)
	ic.Unequip(LegsSlot)
	ic.Unequip(FeetSlot)
}

// EquipBest finds and equips the highest-value item for the given slot.
func (ic *InventoryComponent) EquipBest(slot ItemSlot) {
	ic.Unequip(slot)
	var best *ecs.Entity
	bestVal := -9999
	for _, item := range ic.Bag {
		if !item.HasComponent(Item) {
			continue
		}
		if item.GetComponent(Item).(*ItemComponent).Slot != slot {
			continue
		}
		val := 0
		if item.HasComponent(Weapon) {
			val = item.GetComponent(Weapon).(*WeaponComponent).AttackBonus
		} else if item.HasComponent(Armor) {
			val = item.GetComponent(Armor).(*ArmorComponent).DefenseBonus
		}
		if val > bestVal || (val == bestVal && best != nil && rand.Float64() < 0.5) {
			bestVal = val
			best = item
		}
	}
	if best != nil {
		ic.Equip(best)
	}
}

func (ic *InventoryComponent) EquipAllBest() {
	ic.EquipBest(HandSlot)
	ic.EquipBest(HeadSlot)
	ic.EquipBest(TorsoSlot)
	ic.EquipBest(LegsSlot)
	ic.EquipBest(FeetSlot)
}

func (ic *InventoryComponent) GetAttackModifier() int {
	mod := 0
	for _, hand := range []*ecs.Entity{ic.RightHand, ic.LeftHand} {
		if hand != nil && hand.HasComponent(Weapon) {
			mod += hand.GetComponent(Weapon).(*WeaponComponent).AttackBonus
		}
	}
	return mod
}

func (ic *InventoryComponent) GetAttackDice() string {
	dice := ""
	for _, hand := range []*ecs.Entity{ic.RightHand, ic.LeftHand} {
		if hand != nil && hand.HasComponent(Weapon) {
			d := hand.GetComponent(Weapon).(*WeaponComponent).AttackDice
			if d != "" {
				if dice != "" {
					dice += "+"
				}
				dice += d
			}
		}
	}
	return dice
}

func (ic *InventoryComponent) GetDefenseModifier() int {
	mod := 0
	slots := []*ecs.Entity{ic.Head, ic.Torso, ic.Legs, ic.Feet, ic.LeftHand, ic.RightHand}
	for _, slot := range slots {
		if slot != nil && slot.HasComponent(Armor) {
			mod += slot.GetComponent(Armor).(*ArmorComponent).DefenseBonus
		}
	}
	return mod
}
