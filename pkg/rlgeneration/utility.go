package rlgeneration

import (
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/utility"
)

// BudRooms generates numRooms rooms by scanning for solid tiles adjacent to open
// space and carving a room off them. wallType and floorType are passed through to
// CarveRoom.
//
// canBud is called on a candidate tile to determine whether a room may bud from it.
// Use this to implement the NoBudding flag or any other game-specific restriction.
// Passing nil allows budding from any solid tile.
func BudRooms(level rlworld.LevelInterface, z, width, height, numRooms int, wallType, floorType string, canBud func(rlworld.TileInterface) bool) {
	for i := 0; i < numRooms; i++ {
		done := false
		tries := 0
		for tries < 99999 && !done {
			tries++
			rX := utility.GetRandom(0, width)
			rY := utility.GetRandom(0, height)
			rHeight := utility.GetRandom(4, 10)
			rWidth := utility.GetRandom(4, 10)

			if rX <= 0 || rX >= level.GetWidth() || rY <= 0 || rY >= level.GetHeight() {
				continue
			}

			t := level.GetTileAt(rX, rY, z)
			if t == nil {
				continue
			}
			if canBud != nil && !canBud(t) {
				continue
			}
			if !t.IsSolid() {
				continue
			}

			// Up
			upTile := level.GetTileAt(rX, rY-2, z)
			if upTile != nil && upTile.IsAir() {
				left := level.GetTileAt(rX-1, rY, z)
				right := level.GetTileAt(rX+1, rY, z)
				if left != nil && right != nil && left.IsSolid() && right.IsSolid() {
					if !RoomIntersects(level, rX-rWidth/2, rY-rHeight-1, z, rWidth, rHeight) {
						CarveRoom(level, rX-rWidth/2, rY-rHeight+1, z, rWidth, rHeight, wallType, floorType, false)
						level.UpdateTileAt(rX, rY, z, floorType, 0)
						done = true
					}
				}
			}

			// Down
			if !done {
				downTile := level.GetTileAt(rX, rY+2, z)
				if downTile != nil && downTile.IsAir() {
					left := level.GetTileAt(rX-1, rY, z)
					right := level.GetTileAt(rX+1, rY, z)
					if left != nil && right != nil && left.IsSolid() && right.IsSolid() {
						if !RoomIntersects(level, rX-rWidth/2, rY+1, z, rWidth, rHeight) {
							CarveRoom(level, rX-rWidth/2, rY, z, rWidth, rHeight, wallType, floorType, false)
							level.UpdateTileAt(rX, rY, z, floorType, 0)
							done = true
						}
					}
				}
			}

			// Left
			if !done {
				leftTile := level.GetTileAt(rX-2, rY, z)
				if leftTile != nil && leftTile.IsAir() {
					top := level.GetTileAt(rX, rY-1, z)
					bottom := level.GetTileAt(rX, rY+1, z)
					if top != nil && bottom != nil && top.IsSolid() && bottom.IsSolid() {
						if !RoomIntersects(level, rX-rWidth, rY-rHeight/2, z, rWidth, rHeight) {
							CarveRoom(level, rX-rWidth+1, rY-rHeight/2, z, rWidth, rHeight, wallType, floorType, false)
							level.UpdateTileAt(rX, rY, z, floorType, 0)
							done = true
						}
					}
				}
			}

			// Right
			if !done {
				rightTile := level.GetTileAt(rX+2, rY, z)
				if rightTile != nil && rightTile.IsAir() {
					top := level.GetTileAt(rX, rY-1, z)
					bottom := level.GetTileAt(rX, rY+1, z)
					if top != nil && bottom != nil && top.IsSolid() && bottom.IsSolid() {
						if !RoomIntersects(level, rX+1, rY-rHeight/2, z, rWidth, rHeight) {
							CarveRoom(level, rX, rY-rHeight/2, z, rWidth, rHeight, wallType, floorType, false)
							level.UpdateTileAt(rX, rY, z, floorType, 0)
							done = true
						}
					}
				}
			}
		}
	}
}

// RoomIntersects returns true if the rectangle at (x,y) of the given size is
// out of bounds or overlaps any non-air tile.
func RoomIntersects(level rlworld.LevelInterface, x, y, z, width, height int) bool {
	if x+width > level.GetWidth() || y+height > level.GetHeight() || x < 0 || y < 0 {
		return true
	}
	for cy := 0; cy < height; cy++ {
		for cx := 0; cx < width; cx++ {
			t := level.GetTileAt(x+cx, y+cy, z)
			if t == nil || !t.IsAir() {
				return true
			}
		}
	}
	return false
}

// BuildRecursiveRoom recursively subdivides the rectangle (x1,y1)–(x2,y2) into
// nested rooms and corridors. Ported from ToME2's generate.cc.
func BuildRecursiveRoom(level rlworld.LevelInterface, x1, y1, x2, y2, z, power int, wallType, floorType string) {
	xSize := x2 - x1
	ySize := y2 - y1
	if xSize < 0 || ySize < 0 {
		return
	}

	choice := 0
	if power < 3 && xSize > 12 && ySize > 12 {
		choice = 1
	} else {
		if power < 10 {
			if utility.GetRandom(1, 11) > 2 && xSize < 8 && ySize < 8 {
				choice = 4
			} else {
				choice = utility.GetRandom(1, 3) + 1
			}
		} else {
			choice = utility.GetRandom(1, 4) + 1
		}
	}

	// Outer walls with entrances, then recurse into keep and corners.
	if choice == 1 {
		for x := x1; x <= x2; x++ {
			level.UpdateTileAt(x, y1, z, wallType, 0)
			level.UpdateTileAt(x, y2, z, wallType, 0)
		}
		for y := y1 + 1; y < y2; y++ {
			level.UpdateTileAt(x1, y, z, wallType, 0)
			level.UpdateTileAt(x2, y, z, wallType, 0)
		}
		if utility.GetRandom(0, 2) == 0 {
			y := utility.GetRandom(0, ySize) + y1
			level.UpdateTileAt(x1, y, z, floorType, 0)
			level.UpdateTileAt(x2, y, z, floorType, 0)
		} else {
			x := utility.GetRandom(0, xSize) + x1
			level.UpdateTileAt(x, y1, z, floorType, 0)
			level.UpdateTileAt(x, y2, z, floorType, 0)
		}
		t1 := utility.GetRandom(0, ySize/3) + y1
		t2 := y2 - utility.GetRandom(0, ySize/3)
		t3 := utility.GetRandom(0, xSize/3) + x1
		t4 := x2 - utility.GetRandom(0, xSize/3)
		BuildRecursiveRoom(level, x1+1, y1+1, x2-1, t1, z, power+1, wallType, floorType)
		BuildRecursiveRoom(level, x1+1, t2, x2-1, y2, z, power+1, wallType, floorType)
		BuildRecursiveRoom(level, x1+1, t1+1, t3, t2-1, z, power+3, wallType, floorType)
		BuildRecursiveRoom(level, t4, t1+1, x2-1, t2-1, z, power+3, wallType, floorType)
		x1, x2, y1, y2 = t3, t4, t1, t2
		xSize = x2 - x1
		ySize = y2 - y1
		power += 2
	}

	// Split vertically.
	if choice == 2 {
		if xSize < 3 {
			for y := y1; y < y2; y++ {
				for x := x1; x < x2; x++ {
					level.UpdateTileAt(x, y, z, floorType, 0)
				}
			}
			return
		}
		t1 := utility.GetRandom(1, xSize-2) + x1 + 1
		BuildRecursiveRoom(level, x1, y1, t1, y2, z, power-2, wallType, floorType)
		BuildRecursiveRoom(level, t1+1, y1, x2, y2, z, power-2, wallType, floorType)
	}

	// Split horizontally.
	if choice == 3 {
		if ySize < 3 {
			for y := y1; y < y2; y++ {
				for x := x1; x < x2; x++ {
					level.UpdateTileAt(x, y, z, floorType, 0)
				}
			}
			return
		}
		t1 := utility.GetRandom(1, ySize-2) + y1 + 1
		BuildRecursiveRoom(level, x1, y1, x2, t1, z, power-2, wallType, floorType)
		BuildRecursiveRoom(level, x1, t1+1, x2, y2, z, power-2, wallType, floorType)
	}

	// Create a room.
	if choice == 4 {
		if xSize < 3 || ySize < 3 {
			for y := y1; y < y2; y++ {
				for x := x1; x < x2; x++ {
					level.UpdateTileAt(x, y, z, wallType, 0)
				}
			}
			return
		}
		for x := x1 + 1; x <= x2-1; x++ {
			level.UpdateTileAt(x, y1+1, z, wallType, 0)
			level.UpdateTileAt(x, y2-1, z, wallType, 0)
		}
		for y := y1 + 1; y < y2-1; y++ {
			level.UpdateTileAt(x1+1, y, z, wallType, 0)
			level.UpdateTileAt(x2-1, y, z, wallType, 0)
		}
		y := utility.GetRandom(1, ySize-3) + y1 + 1
		if utility.GetRandom(0, 2) == 0 {
			level.UpdateTileAt(x1+1, y, z, floorType, 0)
		} else {
			level.UpdateTileAt(x2-1, y, z, floorType, 0)
		}
		BuildRecursiveRoom(level, x1+2, y1+2, x2-2, y2-2, z, power+3, wallType, floorType)
	}
}
