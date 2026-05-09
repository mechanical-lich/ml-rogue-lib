package rlworld

// Tile is a GC-friendly tile struct. It stores its flat index and the level
// dimensions needed to derive coordinates — all plain integer fields, so the
// GC never needs to scan a Tile for pointers.
type Tile struct {
	Type       int   `json:"Type"`
	Variant    int   `json:"Variant"`
	LightLevel int   `json:"LightLevel"`
	Radiation  uint8 `json:"Radiation,omitempty"` // 0 = none, 255 = max; tints rendering and damages exposed entities
	Idx        int   `json:"-"`                   // flat index into Level.Data
	width      int   `json:"-"`                   // cached from the owning level
	height     int   `json:"-"`                   // cached from the owning level
}

// Coords derives X, Y, Z from the flat index and the cached level dimensions.
func (t *Tile) Coords() (x, y, z int) {
	x = t.Idx % t.width
	y = (t.Idx / t.width) % t.height
	z = t.Idx / (t.width * t.height)
	return
}

func (t *Tile) IsSolid() bool { return TileDefinitions[t.Type].Solid }
func (t *Tile) IsWater() bool { return TileDefinitions[t.Type].Water }
func (t *Tile) IsAir() bool   { return TileDefinitions[t.Type].Air }

// PathID returns the flat index as a unique node ID for pathfinding.
func (t *Tile) PathID() int { return t.Idx }

// pathOffsets lists the six cardinal directions (4 planar + 2 vertical).
var pathOffsets = [6][3]int{
	{-1, 0, 0}, // Left
	{1, 0, 0},  // Right
	{0, -1, 0}, // Up
	{0, 1, 0},  // Down
	{0, 0, -1}, // Z down
	{0, 0, 1},  // Z up
}

// DefaultPathCost is the base path cost function used by Level.PathCost when
// no custom PathCostFunc is set. It rejects solid/water tiles and enforces
// stair tiles for Z-level transitions.
func DefaultPathCost(from, to *Tile) float64 {
	tileDef := TileDefinitions[to.Type]

	if tileDef.Solid || tileDef.Water {
		return 5000.0
	}

	_, _, fromZ := from.Coords()
	_, _, toZ := to.Coords()

	if fromZ < toZ {
		if !TileDefinitions[from.Type].StairsUp {
			return 1000.0
		}
	} else if fromZ > toZ {
		if !TileDefinitions[from.Type].StairsDown {
			return 1000.0
		}
	}

	return 0.0
}
