package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// NeverSleepComponent causes an entity to always get a turn regardless of time of day.
type NeverSleepComponent struct{}

func (pc NeverSleepComponent) GetType() ecs.ComponentType {
	return NeverSleep
}
