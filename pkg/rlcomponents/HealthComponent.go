package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

type HealthComponent struct {
	MaxHealth int
	Health    int
	Energy    int
}

func (pc HealthComponent) GetType() ecs.ComponentType {
	return Health
}
