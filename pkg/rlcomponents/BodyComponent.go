package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

type BodyPart struct {
	Name                string
	Description         string
	AttachedTo          []string   // Names of body parts this part attaches to (e.g. "head" attaches to "torso")
	HP                  int        // Current HP of this body part
	MaxHP               int        // Maximum HP of this body part
	Broken              bool       // If true, this body part has been broken
	Amputated           bool       // If true, this body part has been amputated
	KillsWhenBroken     bool       // If true, the entity dies when this body part is broken
	KillsWhenAmputated  bool       // If true, the entity dies when this body part is amputated
	CompatibleItemSlots []ItemSlot // Item slots that can be equipped to this body part (e.g. "head" can equip "head" slot items)
}

type BodyComponent struct {
	Parts map[string]BodyPart
}

func (bc BodyComponent) GetType() ecs.ComponentType {
	return Body
}

func (bc *BodyComponent) AddPart(part BodyPart) {
	if bc.Parts == nil {
		bc.Parts = make(map[string]BodyPart)
	}
	bc.Parts[part.Name] = part
}
