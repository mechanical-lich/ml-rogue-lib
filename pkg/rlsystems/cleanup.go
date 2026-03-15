package rlsystems

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
)

// CleanUpSystem removes dead entities from the level each frame.
// It also strips MyTurn from all living entities at the top of each frame.
//
// Extension hooks:
//   - OnEntityDead: called for each dead entity before it is removed.
//     Use this to spawn drops, play death sounds, award XP, etc.
//   - OnEntityRemoved: called immediately after RemoveEntity, for any
//     additional cleanup (e.g., remove from custom component registries).
type CleanUpSystem struct {
	// OnEntityDead is called for each dead entity before removal.
	OnEntityDead func(level rlworld.LevelInterface, entity *ecs.Entity)

	// OnEntityRemoved is called after each entity is removed from the level.
	OnEntityRemoved func(level rlworld.LevelInterface, entity *ecs.Entity)

	// OnEntityCleanup is called for each entity that is cleaned up, regardless of whether it was removed.
	// Use this for cleanup that should happen for all entities, such as removing MyTurn.
	OnEntityCleanup func(level rlworld.LevelInterface, entity *ecs.Entity)

	deadBuf   []*ecs.Entity
	staticBuf []*ecs.Entity
}

// Update runs the cleanup pass. Call once per frame.
func (s *CleanUpSystem) Update(level rlworld.LevelInterface) {
	s.deadBuf = s.deadBuf[:0]
	for _, entity := range level.GetEntities() {
		if entity.HasComponent(rlcomponents.MyTurn) {
			entity.RemoveComponent(rlcomponents.MyTurn)
		}
		if s.OnEntityCleanup != nil {
			s.OnEntityCleanup(level, entity)
		}
		if entity.HasComponent(rlcomponents.Dead) {
			s.deadBuf = append(s.deadBuf, entity)
		}
	}

	for _, entity := range s.deadBuf {
		if s.OnEntityDead != nil {
			s.OnEntityDead(level, entity)
		}

		// Skip removal for food entities with remaining nutrition (corpse still usable).
		if entity.HasComponent(rlcomponents.Food) {
			fc := entity.GetComponent(rlcomponents.Food).(*rlcomponents.FoodComponent)
			if fc.Amount > 0 {
				continue
			}
		}

		level.RemoveEntity(entity)
		if s.OnEntityRemoved != nil {
			s.OnEntityRemoved(level, entity)
		}
	}

	// Clean up dead static entities.
	s.staticBuf = s.staticBuf[:0]
	for _, entity := range level.GetStaticEntities() {
		if entity.HasComponent(rlcomponents.Dead) {
			s.staticBuf = append(s.staticBuf, entity)
		}
	}

	for _, entity := range s.staticBuf {
		if s.OnEntityDead != nil {
			s.OnEntityDead(level, entity)
		}

		if entity.HasComponent(rlcomponents.Food) {
			fc := entity.GetComponent(rlcomponents.Food).(*rlcomponents.FoodComponent)
			if fc.Amount > 0 {
				continue
			}
		}

		level.RemoveEntity(entity)
		if s.OnEntityRemoved != nil {
			s.OnEntityRemoved(level, entity)
		}
	}
}
