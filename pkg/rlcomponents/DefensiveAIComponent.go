package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DefensiveAIComponent causes an entity to retaliate against attackers.
type DefensiveAIComponent struct {
	AttackerX int
	AttackerY int
	Attacked  bool
}

func (pc DefensiveAIComponent) GetType() ecs.ComponentType {
	return DefensiveAI
}
