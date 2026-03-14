package rlgeneration

import (
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/utility"
)

// Room represents an axis-aligned rectangular room by position and size.
type Room struct {
	X, Y, GetWidth, GetHeight int
}

// CarveRoom writes wallType on the border and floorType in the interior of a rectangle.
// When noOverwrite is true, solid tiles are skipped.
func CarveRoom(level rlworld.LevelInterface, x, y, z, width, height int, wallType, floorType string, noOverwrite bool) {
	if x+width > level.GetWidth() || y+height > level.GetHeight() {
		return
	}
	for i := 0; i < width; i++ {
		for j := 0; j < height; j++ {
			if noOverwrite {
				t := level.GetTileAt(x+i, y+j, z)
				if t != nil && t.IsSolid() {
					continue
				}
			}
			if i == 0 || i == width-1 || j == 0 || j == height-1 {
				level.SetTileType(x+i, y+j, wallType)
			} else {
				level.SetTileType(x+i, y+j, floorType)
			}
		}
	}
}

// CarveRect writes wallType on the border and floorType in the interior of a
// rectangle defined by two corners (x1,y1)–(x2,y2).
// When noOverwrite is true, non-solid tiles are skipped (note: original behaviour).
func CarveRect(level rlworld.LevelInterface, x1, y1, x2, y2, z int, wallType, floorType string, noOverwrite bool) {
	if x2 > level.GetWidth() || y2 > level.GetHeight() {
		return
	}
	for x := x1; x <= x2; x++ {
		for y := y1; y <= y2; y++ {
			if noOverwrite {
				t := level.GetTileAt(x, y, z)
				if t != nil && !t.IsSolid() {
					continue
				}
			}
			if x == x1 || x == x2 || y == y1 || y == y2 {
				level.SetTileType(x, y, wallType)
			} else {
				level.SetTileType(x, y, floorType)
			}
		}
	}
}

// CarveCircle carves a circle of radius r centred at (startX, startY).
// The ring at distance >= r-1 receives wallType; the interior receives floorType.
// When noOverwrite is true, solid tiles are skipped.
func CarveCircle(level rlworld.LevelInterface, z, startX, startY, r int, wallType, floorType string, noOverwrite bool) {
	for x := startX - r; x < startX+r; x++ {
		for y := startY - r; y < startY+r; y++ {
			if (x-startX)*(x-startX)+(y-startY)*(y-startY) >= r*r {
				continue
			}
			if noOverwrite {
				t := level.GetTileAt(x, y, z)
				if t != nil && t.IsSolid() {
					continue
				}
			}
			if utility.Distance(x, y, startX, startY) >= r-1 {
				level.SetTileType(x, y, wallType)
			} else {
				level.SetTileType(x, y, floorType)
			}
		}
	}
}
