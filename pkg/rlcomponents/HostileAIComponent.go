package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/path"
)

// HostileAIComponent causes an entity to pursue and attack targets within sight range.
type HostileAIComponent struct {
	SightRange int
	TargetX    int
	TargetY    int
	Path       []path.Pather // cached path to current target
}

func (pc HostileAIComponent) GetType() ecs.ComponentType {
	return HostileAI
}
