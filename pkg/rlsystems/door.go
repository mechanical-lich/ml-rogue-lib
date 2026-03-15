package rlsystems

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// AppearanceUpdater is a minimal interface for the door system to update sprite coordinates.
// Implement this on your Appearance component.
type AppearanceUpdater interface {
	SetSprite(x, y int)
}

// DoorSystem updates the visual sprite of Door entities based on open/closed state.
//
// Extension hook:
//   - OnDoorStateChange: called whenever a door's open/closed state is applied.
//     Use this to play sounds, trigger animations, or update pathfinding caches.
type DoorSystem struct {
	// OnDoorStateChange is called each tick for every door entity.
	// open reflects the current door state.
	OnDoorStateChange func(entity *ecs.Entity, open bool)

	// AppearanceType is the component type for your game's Appearance component.
	// If set, the system will call SetSprite on it via the AppearanceUpdater interface.
	AppearanceType ecs.ComponentType
}

func (s *DoorSystem) Requires() []ecs.ComponentType {
	if s.AppearanceType != "" {
		return []ecs.ComponentType{rlcomponents.Door, s.AppearanceType}
	}
	return []ecs.ComponentType{rlcomponents.Door}
}

func (s *DoorSystem) UpdateSystem(data interface{}) error {
	return nil
}

func (s *DoorSystem) UpdateEntity(levelInterface interface{}, entity *ecs.Entity) error {
	door := entity.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)

	if s.AppearanceType != "" && entity.HasComponent(s.AppearanceType) {
		if ac, ok := entity.GetComponent(s.AppearanceType).(AppearanceUpdater); ok {
			if door.Open {
				ac.SetSprite(door.OpenedSpriteX, door.OpenedSpriteY)
			} else {
				ac.SetSprite(door.ClosedSpriteX, door.ClosedSpriteY)
			}
		}
	}

	if s.OnDoorStateChange != nil {
		s.OnDoorStateChange(entity, door.Open)
	}

	return nil
}
