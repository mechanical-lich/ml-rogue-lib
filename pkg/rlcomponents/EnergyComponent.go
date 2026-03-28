package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// EnergyComponent implements a tick-up action point system.
// Each tick, Energy increases by Speed. The entity can act whenever
// Energy > 0. Actions deduct their cost from Energy, and leftover
// Energy carries over to the next cycle.
type EnergyComponent struct {
	Speed          int // energy gained per tick (higher = faster entity)
	Energy         int // current accumulated energy
	LastActionCost int // energy cost of the most recent action (set by systems, consumed by SpendTurn)
}

func (ec EnergyComponent) GetType() ecs.ComponentType {
	return Energy
}

// CanAct returns true if the entity has enough energy to take an action.
func (ec *EnergyComponent) CanAct() bool {
	return ec.Energy > 0
}

// SpendTurn deducts the cost of the last action from Energy and resets
// LastActionCost. Returns the cost that was deducted.
func (ec *EnergyComponent) SpendTurn() int {
	cost := ec.LastActionCost
	ec.Energy -= cost
	ec.LastActionCost = 0
	return cost
}
