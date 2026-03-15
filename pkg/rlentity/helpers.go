package rlentity

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
)

// GetName returns the entity's Description name, or "Unknown".
func GetName(entity *ecs.Entity) string {
	if entity.HasComponent(rlcomponents.Description) {
		return entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
	}
	return "Unknown"
}

// CanPassThroughDoor returns true if the entity is allowed through the door.
// A door is passable when open, or when the entity's Description faction matches OwnedBy.
func CanPassThroughDoor(entity *ecs.Entity, door *rlcomponents.DoorComponent) bool {
	if door.Open {
		return true
	}
	if entity.HasComponent(rlcomponents.Description) {
		dc := entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
		if dc.Faction == door.OwnedBy && door.OwnedBy != "" {
			return true
		}
	}
	return false
}

// HandleDeath adds DeadComponent when Health reaches zero.
// Returns true if the entity died.
func HandleDeath(entity *ecs.Entity) bool {
	if !entity.HasComponent(rlcomponents.Health) {
		return false
	}
	hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	if hc.Health <= 0 {
		entity.AddComponent(&rlcomponents.DeadComponent{})
		return true
	}
	return false
}

// Face updates the entity's direction component based on movement delta.
func Face(entity *ecs.Entity, deltaX, deltaY int) {
	if !entity.HasComponent(rlcomponents.Direction) {
		return
	}
	dc := entity.GetComponent(rlcomponents.Direction).(*rlcomponents.DirectionComponent)
	if deltaY > 0 {
		dc.Direction = 1
	}
	if deltaY < 0 {
		dc.Direction = 2
	}
	if deltaX < 0 {
		dc.Direction = 3
	}
	if deltaX > 0 {
		dc.Direction = 0
	}
}

// Eat consumes one unit of food from a food entity. Returns true on success.
func Eat(entity, foodEntity *ecs.Entity) bool {
	if entity == foodEntity || !foodEntity.HasComponent(rlcomponents.Food) {
		return false
	}
	fc := foodEntity.GetComponent(rlcomponents.Food).(*rlcomponents.FoodComponent)
	fc.Amount--
	return true
}

// Swap exchanges the positions of two entities on the level.
func Swap(level rlworld.LevelInterface, entity, entityHit *ecs.Entity) {
	if entity == entityHit {
		return
	}
	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	hitPC := entityHit.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	oldX, oldY, oldZ := pc.GetX(), pc.GetY(), pc.GetZ()
	level.PlaceEntity(hitPC.GetX(), hitPC.GetY(), hitPC.GetZ(), entity)
	level.PlaceEntity(oldX, oldY, oldZ, entityHit)
}

// Move attempts to move an entity by (deltaX, deltaY, deltaZ).
// Handles solid entity collisions, door access, and tile property checks.
// Returns true if a solid entity was in the way.
func Move(entity *ecs.Entity, level rlworld.LevelInterface, deltaX, deltaY, deltaZ int) bool {
	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	destX := pc.GetX() + deltaX
	destY := pc.GetY() + deltaY
	destZ := pc.GetZ() + deltaZ

	canMove := true
	blocker := level.GetSolidEntityAt(destX, destY, destZ)
	if blocker != nil {
		canMove = false
		if blocker.HasComponent(rlcomponents.Door) {
			door := blocker.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
			if CanPassThroughDoor(entity, door) {
				canMove = true
			}
		}
	}

	if canMove {
		tile := level.GetTileAt(destX, destY, destZ)
		if tile != nil && !tile.IsSolid() {
			if tile.IsAir() {
				// Stand on top of the solid tile below rather than entering air.
				below := level.GetTileAt(destX, destY, destZ-1)
				if below != nil && below.IsSolid() {
					level.PlaceEntity(destX, destY, destZ, entity)
				}
			} else if !tile.IsWater() {
				level.PlaceEntity(destX, destY, destZ, entity)
			}
		}
		return false
	}
	return true
}

// HandleMovement moves and faces the entity. No-ops if all deltas are zero.
func HandleMovement(level rlworld.LevelInterface, entity *ecs.Entity, deltaX, deltaY, deltaZ int) {
	if deltaX == 0 && deltaY == 0 && deltaZ == 0 {
		return
	}
	Move(entity, level, deltaX, deltaY, deltaZ)
	Face(entity, deltaX, deltaY)
}
