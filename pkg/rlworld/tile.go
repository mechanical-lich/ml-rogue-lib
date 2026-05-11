package rlworld

// TileInterface is the read-only view of a tile used by AI and entity systems.
// Pathfinding is now handled by Level (which implements path.Graph), so only
// the coordinate, identity, and property methods are needed here.
//
// HasFloor reports whether the cell carries its own ground tile. rlworld tiles
// always return false (Minecraft-style: walkability needs a solid tile in the
// cell below). Layered tiles (rllayered) return true when their Floor slot is
// populated; in that case rlentity.Move skips the below-cell solid check.
type TileInterface interface {
	Coords() (x, y, z int)
	PathID() int
	IsSolid() bool
	IsWater() bool
	IsAir() bool
	HasFloor() bool
}
