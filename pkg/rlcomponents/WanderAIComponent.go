package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// WanderAIComponent causes an entity to move randomly each turn.
type WanderAIComponent struct{}

func (pc WanderAIComponent) GetType() ecs.ComponentType {
	return WanderAI
}
