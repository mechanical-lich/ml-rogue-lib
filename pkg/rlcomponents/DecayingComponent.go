package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DecayingComponent is implemented by status effects that expire over time.
// Decay returns true when the effect should be removed.
type DecayingComponent interface {
	Decay() bool
	GetType() ecs.ComponentType
}
