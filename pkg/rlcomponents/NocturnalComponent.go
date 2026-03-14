package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// NocturnalComponent causes an entity to only act at night.
type NocturnalComponent struct{}

func (pc NocturnalComponent) GetType() ecs.ComponentType {
	return Nocturnal
}
