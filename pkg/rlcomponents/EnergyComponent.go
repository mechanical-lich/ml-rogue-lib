package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// EnergyComponent implements a tick-up action point system.
// Each tick, Energy increases by Speed. When Energy reaches the
// Threshold, the entity can act. Actions deduct their cost from Energy,
// and leftover Energy carries over to the next cycle.
type EnergyComponent struct {
	Speed          int // energy gained per tick (higher = faster entity)
	Energy         int // current accumulated energy
	Threshold      int // energy needed to act (e.g. 100)
	LastActionCost int // energy cost of the most recent action (set by systems, consumed by SpendTurn)
}

func (ec EnergyComponent) GetType() ecs.ComponentType {
	return EnergyType
}

// CanAct returns true if the entity has enough energy to take an action.
func (ec *EnergyComponent) CanAct() bool {
	return ec.Energy >= ec.Threshold
}

// SpendTurn deducts the cost of the last action from Energy and resets
// LastActionCost. If no cost was set, Threshold is used as the default.
// Returns the cost that was deducted.
func (ec *EnergyComponent) SpendTurn() int {
	cost := ec.LastActionCost
	if cost == 0 {
		cost = ec.Threshold
	}
	ec.Energy -= cost
	ec.LastActionCost = 0
	return cost
}
