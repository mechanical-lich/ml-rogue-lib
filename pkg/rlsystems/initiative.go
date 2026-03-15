package rlsystems

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
)

// InitiativeSystem ticks entity initiative counters and grants MyTurn when the counter
// reaches zero, respecting nocturnal/diurnal and alert schedules.
//
// Extension hook:
//   - OnEntityTurn: called after MyTurn is added. Use this to trigger custom per-turn
//     setup (e.g., animation bounce, custom state resets).
type InitiativeSystem struct {
	Speed int

	// OnEntityTurn is called each time an entity receives a MyTurn component.
	// Set this to run game-specific logic at turn start.
	OnEntityTurn func(entity *ecs.Entity)
}

var initiativeRequires = []ecs.ComponentType{rlcomponents.Initiative}

func (s *InitiativeSystem) Requires() []ecs.ComponentType {
	return initiativeRequires
}

func (s *InitiativeSystem) UpdateSystem(data interface{}) error {
	return nil
}

func (s *InitiativeSystem) UpdateEntity(levelInterface interface{}, entity *ecs.Entity) error {
	level := levelInterface.(rlworld.LevelInterface)
	ic := entity.GetComponent(rlcomponents.Initiative).(*rlcomponents.InitiativeComponent)
	ic.Ticks -= s.Speed

	if ic.Ticks <= 0 {
		ic.Ticks = ic.DefaultValue
		if ic.OverrideValue > 0 {
			ic.Ticks = ic.OverrideValue
		}

		if !entity.HasComponent(rlcomponents.MyTurn) {
			canGo := false
			if entity.HasComponent(rlcomponents.Nocturnal) {
				canGo = level.IsNight()
			} else {
				canGo = !level.IsNight()
			}
			// Alerted entities override their sleep schedule.
			if entity.HasComponent(rlcomponents.Alerted) {
				canGo = true
			}
			// NeverSleep entities always act.
			if entity.HasComponent(rlcomponents.NeverSleep) {
				canGo = true
			}

			if canGo {
				entity.AddComponent(rlcomponents.GetMyTurn())
				if s.OnEntityTurn != nil {
					s.OnEntityTurn(entity)
				}
			}
		}
	}

	return nil
}
