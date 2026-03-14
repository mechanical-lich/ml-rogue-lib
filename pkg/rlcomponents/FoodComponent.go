package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// FoodComponent marks an entity as edible. Amount tracks remaining nutrition.
type FoodComponent struct {
	Amount int
}

func (pc FoodComponent) GetType() ecs.ComponentType {
	return Food
}
