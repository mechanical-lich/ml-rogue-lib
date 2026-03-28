package rlworld

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
)

// TileDefinition describes one category of tile (e.g. "grass", "stone_wall").
// Games can embed this struct to add domain-specific fields.
type TileDefinition struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Solid       bool          `json:"solid"`
	Water       bool          `json:"water"`
	Door        bool          `json:"door"`
	Air         bool          `json:"air"`
	StairsUp    bool          `json:"stairsUp"`
	StairsDown    bool          `json:"stairsDown"`
	MovementCost  int           `json:"movementCost"`
	AutoTile      int           `json:"autoTile"`
	Variants    []TileVariant `json:"variants"`
}

var (
	// TileDefinitions is the index-based lookup table. Tile.Type is an index into this slice.
	TileDefinitions []TileDefinition
	// TileNameToIndex maps a tile name (e.g. "grass") to its index in TileDefinitions.
	TileNameToIndex map[string]int
	// TileIndexToName maps an index back to a tile name (useful for debugging/serialization).
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

// SetTileDefinitions populates the global registries from a slice of definitions.
// Useful when definitions are built programmatically rather than loaded from JSON.
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
