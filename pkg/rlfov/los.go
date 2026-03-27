package rlfov

import (
	"math"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/utility"
)

// Los reports whether (tX, tY) is visible from (pX, pY) on Z layer z.
// Blocked by solid tiles, door tiles, and closed door entities. Uses Bresenham's line algorithm.
func Los(level *rlworld.Level, pX, pY, tX, tY, z int) bool {
	deltaX := pX - tX
	deltaY := pY - tY

	absDeltaX := math.Abs(float64(deltaX))
	absDeltaY := math.Abs(float64(deltaY))

	signX := utility.Sgn(deltaX)
	signY := utility.Sgn(deltaY)

	if absDeltaX > absDeltaY {
		t := absDeltaY*2 - absDeltaX
		for {
			if t >= 0 {
				tY += signY
				t -= absDeltaX * 2
			}
			tX += signX
			t += absDeltaY * 2
			if tX == pX && tY == pY {
				return true
			}
			tile := level.GetTilePtr(tX, tY, z)
			if tile == nil || tile.IsSolid() {
				return false
			}
			if rlworld.TileDefinitions[tile.Type].Door {
				return false
			}
			if blocksLos(level, tX, tY, z) {
				return false
			}
		}
	}

	t := absDeltaX*2 - absDeltaY
	for {
		if t >= 0 {
			tX += signX
			t -= absDeltaY * 2
		}
		tY += signY
		t += absDeltaX * 2
		if tX == pX && tY == pY {
			return true
		}
		tile := level.GetTilePtr(tX, tY, z)
		if tile == nil || tile.IsSolid() {
			return false
		}
		if rlworld.TileDefinitions[tile.Type].Door {
			return false
		}
		if blocksLos(level, tX, tY, z) {
			return false
		}
	}
}

// blocksLos returns true if a closed door entity exists at (x,y,z).
// Only doors affect LOS; other solid entities (enemies, furniture) do not.
func blocksLos(level *rlworld.Level, x, y, z int) bool {
	e := level.GetSolidEntityAt(x, y, z)
	if e == nil || !e.HasComponent(rlcomponents.Door) {
		return false
	}
	return !e.GetComponent(rlcomponents.Door).(*rlcomponents.DoorComponent).Open
}

// UpdateFieldOfView marks all tiles within radius of (x,y,z) as seen
// if they have line of sight to (x,y,z).
func UpdateFieldOfView(level *rlworld.Level, x, y, z, radius int) {
	for tx := x - radius; tx <= x+radius; tx++ {
		for ty := y - radius; ty <= y+radius; ty++ {
			if !level.InBounds(tx, ty, z) {
				continue
			}
			if Los(level, x, y, tx, ty, z) {
				level.SetSeen(tx, ty, z, true)
			}
		}
	}
}
