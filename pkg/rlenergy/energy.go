package rlenergy

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
)

// ResolveTurn handles end-of-turn bookkeeping for a single entity.
// If the entity has both MyTurn and TurnTaken, it calls SpendTurn on
// the EnergyComponent and removes both markers. Returns true if a turn
// was resolved.
func ResolveTurn(entity *ecs.Entity) bool {
	if !entity.HasComponent(rlcomponents.MyTurn) || !entity.HasComponent(rlcomponents.TurnTaken) {
		return false
	}
	if entity.HasComponent(rlcomponents.Energy) {
		ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		ec.SpendTurn()
	}
	entity.RemoveComponent(rlcomponents.MyTurn)
	entity.RemoveComponent(rlcomponents.TurnTaken)
	return true
}

// CanAct returns true if the entity has an EnergyComponent with enough
// energy to take an action.
func CanAct(entity *ecs.Entity) bool {
	if !entity.HasComponent(rlcomponents.Energy) {
		return false
	}
	ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	return ec.CanAct()
}

// GrantTurn adds MyTurn to the entity if it can act and doesn't already
// have one. Returns true if a turn was granted.
func GrantTurn(entity *ecs.Entity) bool {
	if entity.HasComponent(rlcomponents.MyTurn) {
		return false
	}
	if !CanAct(entity) {
		return false
	}
	entity.AddComponent(rlcomponents.GetMyTurn())
	return true
}

// AdvanceEnergy adds Speed to Energy for every entity that has an
// EnergyComponent. Entities that reach the threshold receive MyTurn.
// Returns (playerGotTurn, anyGotTurn).
func AdvanceEnergy(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool) {
	for _, entity := range entities {
		if !entity.HasComponent(rlcomponents.Energy) {
			continue
		}
		ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		ec.Energy += ec.Speed

		if ec.CanAct() && !entity.HasComponent(rlcomponents.MyTurn) {
			entity.AddComponent(rlcomponents.GetMyTurn())
			entity.RemoveComponent(rlcomponents.TurnTaken)
			anyGotTurn = true
			if entity == player {
				playerGotTurn = true
			}
		}
	}
	return
}

// RegrantTurns re-grants MyTurn to entities that still have enough energy
// after their previous action (multi-action). No energy is ticked — this
// only checks leftover energy. Returns (playerGotTurn, anyGotTurn).
func RegrantTurns(entities []*ecs.Entity, player *ecs.Entity) (playerGotTurn, anyGotTurn bool) {
	for _, entity := range entities {
		if GrantTurn(entity) {
			anyGotTurn = true
			if entity == player {
				playerGotTurn = true
			}
		}
	}
	return
}

// SetActionCost records the energy cost of the action an entity just took.
// The cost is consumed by ResolveTurn or SpendTurn.
func SetActionCost(entity *ecs.Entity, cost int) {
	if entity.HasComponent(rlcomponents.Energy) {
		ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
		ec.LastActionCost = cost
	}
}

// MoveCost returns the energy cost for moving onto the given tile.
// It multiplies baseCost by the tile's MovementCost (treating 0 as 1).
func MoveCost(tile *rlworld.Tile, baseCost int) int {
	def := rlworld.TileDefinitions[tile.Type]
	mult := def.MovementCost
	if mult == 0 {
		mult = 1
	}
	return baseCost * mult
}
