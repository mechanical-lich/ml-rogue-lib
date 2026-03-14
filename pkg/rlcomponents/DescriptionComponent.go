package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DescriptionComponent holds an entity's display name and faction affiliation.
type DescriptionComponent struct {
	Name    string
	Faction string
}

func (pc DescriptionComponent) GetType() ecs.ComponentType {
	return Description
}
