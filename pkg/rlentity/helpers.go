package rlentity

import (
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
)

func Face(entity *ecs.Entity, deltaX int, deltaY int) {
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

func Swap(level rlworld.LevelInterface, entity *ecs.Entity, entityHit *ecs.Entity) {
	if entityHit != entity {
		pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		hitPc := entityHit.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

		oldX := pc.GetX()
		oldY := pc.GetY()
		oldZ := pc.GetZ()

		level.PlaceEntity(hitPc.GetX(), hitPc.GetY(), hitPc.GetZ(), entity)
		level.PlaceEntity(oldX, oldY, oldZ, entityHit)
	}
}
