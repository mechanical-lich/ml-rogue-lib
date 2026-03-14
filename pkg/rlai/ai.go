package rlai

import (
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/path"
)

// TrackTarget returns the unit delta needed to move from (x,y,z) toward (x2,y2,z2).
func TrackTarget(x, y, z, x2, y2, z2 int) (int, int, int) {
	dx, dy, dz := 0, 0, 0
	if x < x2 {
		dx = 1
	} else if x > x2 {
		dx = -1
	}
	if y < y2 {
		dy = 1
	} else if y > y2 {
		dy = -1
	}
	if z < z2 {
		dz = 1
	} else if z > z2 {
		dz = -1
	}
	return dx, dy, dz
}

// WithinRange returns true if (x,y,z) is within the given range of (x2,y2,z2).
func WithinRange(x, y, z, x2, y2, z2, rangeX, rangeY, rangeZ int) bool {
	return x >= x2-rangeX && x <= x2+rangeX &&
		y >= y2-rangeY && y <= y2+rangeY &&
		z >= z2-rangeZ && z <= z2+rangeZ
}

// WithinRangeCardinal returns true if (x,y) is within range cardinally (no diagonals).
func WithinRangeCardinal(x, y, x2, y2, rangeX, rangeY int) bool {
	if x == x2 && y >= y2-rangeY && y <= y2+rangeY {
		return true
	}
	if y == y2 && x >= x2-rangeX && x <= x2+rangeX {
		return true
	}
	return false
}

// MoveTowardsTarget moves entity one step along a cached path toward (targetX,targetY,targetZ).
// getPath is called to (re-)compute the path when needed.
// Returns true if a step was taken.
func MoveTowardsTarget(
	level rlworld.LevelInterface,
	entity *ecs.Entity,
	targetX, targetY, targetZ int,
	getPath func(level rlworld.LevelInterface, from, to rlworld.TileInterface, reuse []path.Pather) []path.Pather,
) bool {
	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	mem := entity.GetComponent(rlcomponents.AIMemory).(*rlcomponents.AIMemoryComponent)

	needNew := len(mem.CurrentSteps) < 2 ||
		mem.TargetX != targetX || mem.TargetY != targetY || mem.TargetZ != targetZ

	if needNew {
		from := level.GetTileAt(pc.GetX(), pc.GetY(), pc.GetZ())
		to := level.GetTileAt(targetX, targetY, targetZ)
		if from == nil || to == nil {
			return false
		}
		mem.CurrentSteps = getPath(level, from, to, mem.CurrentSteps)
		mem.TargetX = targetX
		mem.TargetY = targetY
		mem.TargetZ = targetZ
	}

	for len(mem.CurrentSteps) > 1 {
		next := mem.CurrentSteps[1].(rlworld.TileInterface)
		nx, ny, nz := next.Coords()
		if pc.GetX() == nx && pc.GetY() == ny && pc.GetZ() == nz {
			mem.CurrentSteps = mem.CurrentSteps[1:]
			continue
		}
		if next.IsSolid() || level.GetSolidEntityAt(nx, ny, nz) != nil {
			mem.CurrentSteps = nil
			return false
		}
		dx, dy, dz := TrackTarget(pc.GetX(), pc.GetY(), pc.GetZ(), nx, ny, nz)
		rlentity.HandleMovement(level, entity, dx, dy, dz)
		return true
	}
	return false
}
