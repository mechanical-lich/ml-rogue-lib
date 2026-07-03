package rllayered

import "github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"

// Tile is a per-cell record carrying two independently-painted layer slots
// (Floor and Middle) plus runtime fields (lighting, radiation). It is
// GC-friendly: every field is a plain integer so the GC never needs to scan a
// Tile for pointers.
//
// There is no ceiling slot: the "ceiling" of a cell is simply the Floor of the
// cell directly above it (Dwarf-Fortress style). This keeps a single source of
// truth for each horizontal surface, so digging up/down clears one Floor slot
// rather than a floor and a redundant ceiling.
//
// *Tile implements rlworld.TileInterface (Coords, PathID, IsSolid, IsWater,
// IsAir), so existing rlworld-typed callers can consume layered tiles
// through the shared interface.
type Tile struct {
	Floor      Slot  `json:"f"`
	Middle     Slot  `json:"m"`
	LightLevel int   `json:"l"`
	Radiation  uint8 `json:"r,omitempty"`
	Idx        int   `json:"-"`
	width      int   `json:"-"`
	height     int   `json:"-"`
}

// Coords derives X, Y, Z from the flat index and the cached level dimensions.
func (t *Tile) Coords() (x, y, z int) {
	x = t.Idx % t.width
	y = (t.Idx / t.width) % t.height
	z = t.Idx / (t.width * t.height)
	return
}

// PathID returns the flat index as a unique node ID for pathfinding.
func (t *Tile) PathID() int { return t.Idx }

// MiddleDef returns the TileDefinition for the Middle slot. Returns nil if
// Middle is empty.
func (t *Tile) MiddleDef() *TileDefinition {
	if t.Middle.IsEmpty() {
		return nil
	}
	return &TileDefinitions[t.Middle.Type]
}

// FloorDef returns the TileDefinition for the Floor slot. Returns nil if
// Floor is empty.
func (t *Tile) FloorDef() *TileDefinition {
	if t.Floor.IsEmpty() {
		return nil
	}
	return &TileDefinitions[t.Floor.Type]
}

// IsSolid reports whether the Middle slot blocks movement (legacy semantics —
// callers wanting layered walkability should use Level.IsWalkable instead).
func (t *Tile) IsSolid() bool {
	if t.Middle.IsEmpty() {
		return false
	}
	return TileDefinitions[t.Middle.Type].Solid
}

// IsWater reports whether the Middle slot is water.
func (t *Tile) IsWater() bool {
	if t.Middle.IsEmpty() {
		return false
	}
	return TileDefinitions[t.Middle.Type].Water
}

// IsAir reports whether the Middle slot is air (or empty).
func (t *Tile) IsAir() bool {
	if t.Middle.IsEmpty() {
		return true
	}
	return TileDefinitions[t.Middle.Type].Air
}

// HasFloor reports whether the Floor slot is populated — i.e. the cell carries
// its own ground. rlentity.Move uses this to skip the air-over-solid check.
func (t *Tile) HasFloor() bool {
	return !t.Floor.IsEmpty()
}

// Compile-time check that *Tile satisfies rlworld.TileInterface so that
// rllayered.Level can return tiles to rlworld-typed consumers.
var _ rlworld.TileInterface = (*Tile)(nil)

// pathOffsets lists the six cardinal directions (4 planar + 2 vertical).
var pathOffsets = [6][3]int{
	{-1, 0, 0},
	{1, 0, 0},
	{0, -1, 0},
	{0, 1, 0},
	{0, 0, -1},
	{0, 0, 1},
}

// DefaultPathCost rejects cells with no floor (nothing to stand on), tiles
// whose Middle slot is solid/water, and z-transitions that don't originate
// from a stair tile.
func DefaultPathCost(from, to *Tile) float64 {
	if to.Floor.IsEmpty() {
		return 5000.0
	}
	if !to.Middle.IsEmpty() {
		mid := TileDefinitions[to.Middle.Type]
		if mid.Solid || mid.Water {
			return 5000.0
		}
	}

	_, _, fromZ := from.Coords()
	_, _, toZ := to.Coords()

	if fromZ != toZ {
		fromMid := from.Middle
		if fromMid.IsEmpty() {
			return 1000.0
		}
		def := TileDefinitions[fromMid.Type]
		if fromZ < toZ && !def.StairsUp {
			return 1000.0
		}
		if fromZ > toZ && !def.StairsDown {
			return 1000.0
		}
	}

	return 0.0
}
