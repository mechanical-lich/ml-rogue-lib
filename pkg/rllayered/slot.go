package rllayered

// TileLayer names the slots a layered cell can hold. There is no ceiling layer:
// a cell's ceiling is the Floor of the cell above it (see Tile).
type TileLayer int

const (
	LayerMiddle TileLayer = iota // structural / blocking — walls, ore, doors, air, water
	LayerFloor                   // the ground surface
)

// Slot holds a single tile's identity within a layer. Type==0 is the sentinel
// "empty" — the layer carries no tile.
type Slot struct {
	Type    int `json:"t"`
	Variant int `json:"v"`
}

// IsEmpty reports whether the slot is unfilled.
func (s Slot) IsEmpty() bool { return s.Type == 0 }
