package rllayered

// TileLayer names the three slots a layered cell can hold.
type TileLayer int

const (
	LayerMiddle  TileLayer = iota // structural / blocking — walls, ore, doors, air, water
	LayerFloor                    // the ground surface
	LayerCeiling                  // overhead — roofs, canopies (unused in POC)
)

// Slot holds a single tile's identity within a layer. Type==0 is the sentinel
// "empty" — the layer carries no tile.
type Slot struct {
	Type    int `json:"t"`
	Variant int `json:"v"`
}

// IsEmpty reports whether the slot is unfilled.
func (s Slot) IsEmpty() bool { return s.Type == 0 }
