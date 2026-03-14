package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// LightSensitiveComponent marks an entity that catches fire when exposed to bright light.
type LightSensitiveComponent struct{}

func (pc LightSensitiveComponent) GetType() ecs.ComponentType {
	return LightSensitive
}
