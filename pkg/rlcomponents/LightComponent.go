package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// LightComponent causes an entity to emit light in a radius.
type LightComponent struct {
	Level int // brightness (0-100)
	Range int // radius in tiles
}

func (pc LightComponent) GetType() ecs.ComponentType {
	return Light
}
