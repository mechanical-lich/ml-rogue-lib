package rllayered

import (
	"encoding/json"
	"os"
)

// TileVariant holds sprite-sheet coordinates for a visual variant of a tile.
type TileVariant struct {
	Variant int `json:"variant"`
	SpriteX int `json:"spriteX"`
	SpriteY int `json:"spriteY"`
}

// AutoTile modes control how ResolveVariant selects the visual variant for a tile.
const (
	AutoTileNone    = 0 // Use tile.Variant as-is (default)
	AutoTileWall    = 1 // 2-variant: connected bottom → Variants[0], edge → Variants[1]
	AutoTileBitmask = 2 // 4-bit cardinal bitmask (top|bottom|left|right) → 16 variants
	AutoTileBlob47  = 3 // 8-direction with corner pruning → 47 variants (GMS2-style)
)

// TileDefinition describes one category of tile. The Layer field decides which
// slot of a layered cell receives this tile when PaintTile is called by name.
type TileDefinition struct {
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Layer         string        `json:"layer"` // "floor" | "middle" | "ceiling"; default "middle"
	Solid         bool          `json:"solid"`
	Water         bool          `json:"water"`
	Door          bool          `json:"door"`
	Air           bool          `json:"air"`
	StairsUp      bool          `json:"stairsUp"`
	StairsDown    bool          `json:"stairsDown"`
	MovementCost  int           `json:"movementCost"`
	AutoTile      int           `json:"autoTile"`
	Variants      []TileVariant `json:"variants"`
	Resource      string        `json:"resource"`
	SpriteWidth   int           `json:"spriteWidth"`
	SpriteHeight  int           `json:"spriteHeight"`
	SpriteOffsetX int           `json:"spriteOffsetX"`
	SpriteOffsetY int           `json:"spriteOffsetY"`
	LightLevel    int           `json:"lightLevel"`
	LightRange    int           `json:"lightRange"`

	// MixID groups visually-fusing tiles. When two adjacent tiles share the
	// same non-zero MixID, autotile treats them as the same neighbor so the
	// boundary between them doesn't produce inward-edge pieces. 0 = no group.
	MixID int `json:"mixId,omitempty"`
}

// LayerOf returns the slot this tile definition wants to be painted into.
func (d *TileDefinition) LayerOf() TileLayer {
	switch d.Layer {
	case "floor":
		return LayerFloor
	default:
		return LayerMiddle
	}
}

var (
	TileDefinitions []TileDefinition
	TileNameToIndex map[string]int
	TileIndexToName []string
)

// LoadTileDefinitions reads a JSON array of TileDefinition from path and populates the global registries.
func LoadTileDefinitions(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var defs []TileDefinition
	if err := json.NewDecoder(file).Decode(&defs); err != nil {
		return err
	}

	SetTileDefinitions(defs)
	return nil
}

// SetTileDefinitions populates the global registries. Index 0 is reserved as
// the sentinel "empty" slot; callers should ensure their first definition is
// either a real "empty" tile or otherwise unused.
func SetTileDefinitions(defs []TileDefinition) {
	TileDefinitions = make([]TileDefinition, len(defs))
	TileNameToIndex = make(map[string]int, len(defs))
	TileIndexToName = make([]string, len(defs))
	for i, def := range defs {
		TileDefinitions[i] = def
		TileNameToIndex[def.Name] = i
		TileIndexToName[i] = def.Name
	}
}
