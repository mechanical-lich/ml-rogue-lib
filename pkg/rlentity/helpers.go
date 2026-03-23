package rlentity

import (
	"math/rand"
	"slices"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
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

	if door.Locked {
		if door.KeyId != "" {
			if entity.HasComponent(rlcomponents.Inventory) {
				inv := entity.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent)
				for _, item := range inv.Bag {
					if item.HasComponent(rlcomponents.Key) {
						key := item.GetComponent(rlcomponents.Key).(*rlcomponents.KeyComponent)
						if key.KeyID == door.KeyId {
							door.Locked = false
							door.Open = true
							return true
						}
					}
				}
			}
		}
		return false
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

	w, h := 1, 1
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		if sc.Width > 0 {
			w = sc.Width
		}
		if sc.Height > 0 {
			h = sc.Height
		}
	}
	startX := destX - w/2
	startY := destY - h/2

	// Check every footprint tile for entity blockers.
	entityBlocked := false
	for dx := 0; dx < w && !entityBlocked; dx++ {
		for dy := 0; dy < h && !entityBlocked; dy++ {
			blocker := level.GetSolidEntityAt(startX+dx, startY+dy, destZ)
			if blocker != nil && blocker != entity {
				entityBlocked = true
				if blocker.HasComponent(rlcomponents.Door) {
					door := blocker.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent)
					if CanPassThroughDoor(entity, door) {
						entityBlocked = false
					}
				}
			}
		}
	}
	if entityBlocked {
		return true
	}

	// Check every footprint tile for passable terrain.
	for dx := 0; dx < w; dx++ {
		for dy := 0; dy < h; dy++ {
			tx, ty := startX+dx, startY+dy
			tile := level.GetTileAt(tx, ty, destZ)
			if tile == nil || tile.IsSolid() || tile.IsWater() {
				return false
			}
			if tile.IsAir() {
				below := level.GetTileAt(tx, ty, destZ-1)
				if below == nil || !below.IsSolid() {
					return false
				}
			}
		}
	}

	level.PlaceEntity(destX, destY, destZ, entity)
	return false
}

// FootprintBlockers appends all solid entities that overlap with entity's
// footprint at (destX, destY, destZ) to buf, excluding entity itself.
// Useful for sized entities to find everything they'd bump into.
func FootprintBlockers(entity *ecs.Entity, level rlworld.LevelInterface, destX, destY, destZ int, buf *[]*ecs.Entity) {
	w, h := 1, 1
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		if sc.Width > 0 {
			w = sc.Width
		}
		if sc.Height > 0 {
			h = sc.Height
		}
	}
	startX := destX - w/2
	startY := destY - h/2
	*buf = (*buf)[:0]
	seen := map[*ecs.Entity]bool{}
	for dx := 0; dx < w; dx++ {
		for dy := 0; dy < h; dy++ {
			blocker := level.GetSolidEntityAt(startX+dx, startY+dy, destZ)
			if blocker != nil && blocker != entity && !seen[blocker] {
				seen[blocker] = true
				*buf = append(*buf, blocker)
			}
		}
	}
}

// HandleMovement moves and faces the entity. No-ops if all deltas are zero.
func HandleMovement(level rlworld.LevelInterface, entity *ecs.Entity, deltaX, deltaY, deltaZ int) {
	if deltaX == 0 && deltaY == 0 && deltaZ == 0 {
		return
	}
	Move(entity, level, deltaX, deltaY, deltaZ)
	Face(entity, deltaX, deltaY)
}

// CheckExcuseMe posts a random ExcuseMeAnnouncement from bumped when bumper
// collides with it and they swap positions. Call this after a friendly swap.
func CheckExcuseMe(bumped *ecs.Entity) {
	if bumped == nil || !bumped.HasComponent(rlcomponents.Description) {
		return
	}
	dc := bumped.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
	if len(dc.ExcuseMeAnnouncements) == 0 {
		return
	}
	msg := dc.ExcuseMeAnnouncements[rand.Intn(len(dc.ExcuseMeAnnouncements))]
	message.PostTaggedMessage("excuseme", dc.Name, msg)
}

// CheckInteraction fires all triggers on target's InteractionComponent if it
// has one and has not been used (or is repeatable). Posts one InteractionEvent
// per trigger to mlge's queued event manager. Returns true if it fired.
func CheckInteraction(actor, target *ecs.Entity) bool {
	if target == nil || !target.HasComponent(rlcomponents.Interaction) {
		return false
	}
	ic := target.GetComponent(rlcomponents.Interaction).(*rlcomponents.InteractionComponent)
	if ic.Used && !ic.Repeatable {
		return false
	}
	if len(ic.Prompt) > 0 {
		message.PostTaggedMessage("interaction", GetName(target), ic.Prompt)
	}
	for _, trigger := range ic.Triggers {
		event.GetQueuedInstance().QueueEvent(rlcomponents.InteractionEvent{
			Actor:   actor,
			Target:  target,
			Trigger: trigger,
		})
	}
	if !ic.Repeatable {
		ic.Used = true
	}
	return true
}

// FindByID searches level entities for the first one whose DescriptionComponent
// ID matches the given string. Returns nil if not found.
func FindByID(level rlworld.LevelInterface, id string) *ecs.Entity {
	if id == "" {
		return nil
	}
	for _, e := range level.GetEntities() {
		if e == nil || !e.HasComponent(rlcomponents.Description) {
			continue
		}
		dc := e.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
		if dc.ID == id {
			return e
		}
	}
	return nil
}

// FindByTag returns all entities whose DescriptionComponent Tags contain tag.
func FindByTag(level rlworld.LevelInterface, tag string) []*ecs.Entity {
	var results []*ecs.Entity
	for _, e := range level.GetEntities() {
		if e == nil || !e.HasComponent(rlcomponents.Description) {
			continue
		}
		dc := e.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
		if slices.Contains(dc.Tags, tag) {
			results = append(results, e)
		}
	}
	return results
}

// CheckPassOver is called after a successful Move. For each entity at (x, y, z)
// other than the mover it:
//   - fires InteractionComponent triggers (walk-over pressure plates, floor triggers, etc.)
//   - posts a random PassOverDescription message (first match only)
func CheckPassOver(mover *ecs.Entity, level rlworld.LevelInterface, x, y, z int) {
	var buf []*ecs.Entity
	level.GetEntitiesAt(x, y, z, &buf)
	passOverPosted := false
	for _, e := range buf {
		if e == mover {
			continue
		}
		CheckInteraction(mover, e)
		if !passOverPosted && e.HasComponent(rlcomponents.Description) {
			dc := e.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
			if len(dc.PassOverDescription) > 0 {
				msg := dc.PassOverDescription[rand.Intn(len(dc.PassOverDescription))]
				message.PostLocatedTaggedMessage("passover", dc.Name, msg, x, y, z)
				passOverPosted = true
			}
		}
	}
}

// CheckDeathAnnouncement posts a random DeathAnnouncement from the dying entity
// if watcher has line-of-sight to it. Call this before the entity is removed.
func CheckDeathAnnouncement(watcher *ecs.Entity, dying *ecs.Entity, level *rlworld.Level) {
	if watcher == nil || watcher == dying {
		return
	}
	if !dying.HasComponent(rlcomponents.Description) {
		return
	}
	dc := dying.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
	var msg string
	if len(dc.DeathAnnouncements) == 0 {
		msg = dc.Name + " has died."
	} else {
		msg = dc.DeathAnnouncements[rand.Intn(len(dc.DeathAnnouncements))]
	}
	if !dying.HasComponent(rlcomponents.Position) || !watcher.HasComponent(rlcomponents.Position) {
		return
	}
	dp := dying.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	wp := watcher.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	if dp.GetZ() != wp.GetZ() {
		return
	}
	if !rlfov.Los(level, wp.GetX(), wp.GetY(), dp.GetX(), dp.GetY(), dp.GetZ()) {
		return
	}
	message.PostLocatedTaggedMessage("death", dc.Name, msg, dp.GetX(), dp.GetY(), dp.GetZ())
}
