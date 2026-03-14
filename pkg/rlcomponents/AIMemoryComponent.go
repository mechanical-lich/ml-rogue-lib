package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/path"
)

// MemoryFact stores a single remembered fact about another entity.
type MemoryFact struct {
	Key    string
	Name   string // name of the entity this memory is about
	Effect string // e.g., "resists fire", "weak to poison"
	Action string // what action was taken
}

// AIMemoryComponent is a general-purpose state machine and memory store for AI entities.
// It tracks current state, movement target, cached path, and combat history.
type AIMemoryComponent struct {
	Memory             map[string][]MemoryFact
	AttackerX          int
	AttackerY          int
	Attacked           bool
	State              string
	TargetX            int
	TargetY            int
	TargetZ            int
	CurrentSteps       []path.Pather
	FoodSearchCooldown int
}

func (c *AIMemoryComponent) GetType() ecs.ComponentType {
	return AIMemory
}
