package rlcomponents

import (
	"math/rand/v2"

	"github.com/mechanical-lich/mlge/ecs"
)

// BodyInventoryComponent is an inventory whose equipment slots map directly to
// BodyPart names. A part accepts an item when the item's Slot is listed in the
// part's CompatibleItemSlots. Use alongside BodyComponent; the existing
// InventoryComponent remains for entities without body parts.
type BodyInventoryComponent struct {
	Equipped          map[string]*ecs.Entity // keyed by BodyPart.Name
	Bag               []*ecs.Entity
	StartingInventory []string
}

func (ic BodyInventoryComponent) GetType() ecs.ComponentType {
	return BodyInventory
}

// --- Bag operations (mirror InventoryComponent) ---

func (ic *BodyInventoryComponent) AddItem(item *ecs.Entity) {
	ic.Bag = append(ic.Bag, item)
}

func (ic *BodyInventoryComponent) RemoveItem(item *ecs.Entity) bool {
	for i, v := range ic.Bag {
		if v == item {
			ic.Bag = append(ic.Bag[:i], ic.Bag[i+1:]...)
			return true
		}
	}
	return false
}

func (ic *BodyInventoryComponent) RemoveItemByName(name string) bool {
	for i := len(ic.Bag) - 1; i >= 0; i-- {
		if ic.Bag[i].Blueprint == name {
			ic.Bag = append(ic.Bag[:i], ic.Bag[i+1:]...)
			return true
		}
	}
	return false
}

func (ic *BodyInventoryComponent) RemoveAll(name string) bool {
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

func (ic *BodyInventoryComponent) HasItem(name string) bool {
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

// --- Equip operations ---

func (ic *BodyInventoryComponent) initEquipped() {
	if ic.Equipped == nil {
		ic.Equipped = make(map[string]*ecs.Entity)
	}
}

// EquipToBodyPart places item in the named body part slot. Any previously
// equipped item is returned to the bag.
func (ic *BodyInventoryComponent) EquipToBodyPart(item *ecs.Entity, partName string) {
	ic.initEquipped()
	if existing := ic.Equipped[partName]; existing != nil {
		ic.AddItem(existing)
	}
	ic.Equipped[partName] = item
	ic.RemoveItem(item)
}

// AutoEquip finds the first compatible, non-amputated body part for the item's
// slot. Prefers unoccupied parts; displaces an occupied one if none are free.
// Returns true if the item was equipped.
func (ic *BodyInventoryComponent) AutoEquip(item *ecs.Entity, bc *BodyComponent) bool {
	if !item.HasComponent(Item) {
		return false
	}
	slot := item.GetComponent(Item).(*ItemComponent).Slot

	// Prefer empty slots first.
	for partName, part := range bc.Parts {
		if part.Amputated || !partAcceptsSlot(part, slot) {
			continue
		}
		if ic.Equipped[partName] == nil {
			ic.EquipToBodyPart(item, partName)
			return true
		}
	}
	// Fall back: displace the first compatible slot.
	for partName, part := range bc.Parts {
		if part.Amputated || !partAcceptsSlot(part, slot) {
			continue
		}
		ic.EquipToBodyPart(item, partName)
		return true
	}
	return false
}

// Unequip removes the item from the named body part and returns it to the bag.
// Returns the unequipped item, or nil if the slot was empty.
func (ic *BodyInventoryComponent) Unequip(partName string) *ecs.Entity {
	item := ic.Equipped[partName]
	if item != nil {
		delete(ic.Equipped, partName)
		ic.AddItem(item)
	}
	return item
}

// UnequipAll returns all equipped items to the bag.
func (ic *BodyInventoryComponent) UnequipAll() {
	for partName := range ic.Equipped {
		ic.Unequip(partName)
	}
}

// HandleAmputation unequips any item held by the amputated part and returns it
// to the bag. Returns the displaced item, or nil.
func (ic *BodyInventoryComponent) HandleAmputation(partName string) *ecs.Entity {
	return ic.Unequip(partName)
}

// --- Best-fit equipping ---

// EquipBest unequips and re-equips the highest-value item for the given slot
// across every compatible, non-amputated body part.
func (ic *BodyInventoryComponent) EquipBest(slot ItemSlot, bc *BodyComponent) {
	for partName, part := range bc.Parts {
		if part.Amputated || !partAcceptsSlot(part, slot) {
			continue
		}
		ic.Unequip(partName)
		if best := ic.bestItemForSlot(slot); best != nil {
			ic.EquipToBodyPart(best, partName)
		}
	}
}

// EquipAllBest calls EquipBest for every unique slot accepted by any body part.
func (ic *BodyInventoryComponent) EquipAllBest(bc *BodyComponent) {
	seen := map[ItemSlot]bool{}
	for _, part := range bc.Parts {
		for _, slot := range part.CompatibleItemSlots {
			if !seen[slot] {
				seen[slot] = true
				ic.EquipBest(slot, bc)
			}
		}
	}
}

// --- Combat stats ---

// GetAttackModifier returns the total attack bonus from all equipped weapons.
func (ic *BodyInventoryComponent) GetAttackModifier() int {
	mod := 0
	for _, item := range ic.Equipped {
		if item != nil && item.HasComponent(Weapon) {
			mod += item.GetComponent(Weapon).(*WeaponComponent).AttackBonus
		}
	}
	return mod
}

// GetAttackDice returns a combined dice string for all equipped weapons.
func (ic *BodyInventoryComponent) GetAttackDice() string {
	result := ""
	for _, item := range ic.Equipped {
		if item == nil || !item.HasComponent(Weapon) {
			continue
		}
		d := item.GetComponent(Weapon).(*WeaponComponent).AttackDice
		if d != "" {
			if result != "" {
				result += "+"
			}
			result += d
		}
	}
	return result
}

// GetDefenseModifier returns the total defense bonus from all equipped armor.
func (ic *BodyInventoryComponent) GetDefenseModifier() int {
	mod := 0
	for _, item := range ic.Equipped {
		if item != nil && item.HasComponent(Armor) {
			mod += item.GetComponent(Armor).(*ArmorComponent).DefenseBonus
		}
	}
	return mod
}

// GetDamageType returns the damage type of the first equipped weapon, or "".
func (ic *BodyInventoryComponent) GetDamageType() string {
	for _, item := range ic.Equipped {
		if item != nil && item.HasComponent(Weapon) {
			if dt := item.GetComponent(Weapon).(*WeaponComponent).DamageType; dt != "" {
				return dt
			}
		}
	}
	return ""
}

// GetResistances returns all resistance damage types from equipped armor.
func (ic *BodyInventoryComponent) GetResistances() []string {
	var out []string
	for _, item := range ic.Equipped {
		if item != nil && item.HasComponent(Armor) {
			out = append(out, item.GetComponent(Armor).(*ArmorComponent).Resistances...)
		}
	}
	return out
}

// --- Private helpers ---

func partAcceptsSlot(part BodyPart, slot ItemSlot) bool {
	for _, s := range part.CompatibleItemSlots {
		if s == slot {
			return true
		}
	}
	return false
}

func (ic *BodyInventoryComponent) bestItemForSlot(slot ItemSlot) *ecs.Entity {
	var best *ecs.Entity
	bestVal := -9999
	for _, item := range ic.Bag {
		if !item.HasComponent(Item) {
			continue
		}
		if item.GetComponent(Item).(*ItemComponent).Slot != slot {
			continue
		}
		val := bodyInvItemValue(item)
		if val > bestVal || (val == bestVal && best != nil && rand.Float64() < 0.5) {
			bestVal = val
			best = item
		}
	}
	return best
}

func bodyInvItemValue(item *ecs.Entity) int {
	if item.HasComponent(Weapon) {
		return item.GetComponent(Weapon).(*WeaponComponent).AttackBonus
	}
	if item.HasComponent(Armor) {
		return item.GetComponent(Armor).(*ArmorComponent).DefenseBonus
	}
	return 0
}
