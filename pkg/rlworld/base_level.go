package rlworld

import (
	"errors"
	"log"
	"runtime"
	"sync"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/utility"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
)

// Level is a GC-optimized 3D tile container with spatial entity indexing.
// Games can embed this struct and extend it with rendering, lighting overlays, etc.
type Level struct {
	Data           []Tile
	Seen           []bool // parallel to Data — fog of war explored state (1 byte per tile)
	Entities       []*ecs.Entity
	StaticEntities []*ecs.Entity
	entityPos      map[int][]*ecs.Entity
	Width          int
	Height         int
	Depth          int
	Hour           int
	Day            int

	// DirtyColumns tracks (x,y) columns where terrain changed (flat index = y*Width+x).
	DirtyColumns []int

	// PathCostFunc is called by Tile.PathNeighborCost to determine movement cost.
	// Set this to inject game-specific logic (entity blocking, doors, factions, etc.).
	// If nil, DefaultPathCost is used.
	PathCostFunc func(from, to *Tile) float64
}

// Compile-time check that *Level implements LevelInterface.
var _ LevelInterface = (*Level)(nil)

// NewLevel creates a Level with the given dimensions and initializes all tiles to "air".
func NewLevel(width, height, depth int) *Level {
	total := width * height * depth
	level := &Level{
		Width: width, Height: height, Depth: depth,
		Hour:      10,
		Data:      make([]Tile, total),
		Seen:      make([]bool, total),
		entityPos: make(map[int][]*ecs.Entity, 2048),
	}
	level.InitTiles()
	return level
}

// InitTiles initializes all tiles to air in parallel across available CPUs.
func (level *Level) InitTiles() {
	log.Println("Initializing tiles")
	numWorkers := runtime.NumCPU() - 1
	if numWorkers < 1 {
		numWorkers = 1
	}
	totalTiles := level.Width * level.Height * level.Depth
	chunkSize := (totalTiles + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		start := w * chunkSize
		end := (w + 1) * chunkSize
		if end > totalTiles {
			end = totalTiles
		}
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				level.Data[i] = Tile{
					Type:   TileNameToIndex["air"],
					Idx:    i,
					width:  level.Width,
					height: level.Height,
				}
			}
		}(start, end)
	}
	wg.Wait()
}

// ─── Dimensions & bounds ─────────────────────────────────────────────

func (level *Level) GetWidth() int  { return level.Width }
func (level *Level) GetHeight() int { return level.Height }
func (level *Level) GetDepth() int  { return level.Depth }

func (level *Level) InBounds(x, y, z int) bool {
	return x >= 0 && y >= 0 && z >= 0 && x < level.Width && y < level.Height && z < level.Depth
}

// AreNeighborsTheSame checks the four cardinal neighbors of t on the same Z level
// and returns whether each shares the same Type and Variant (useful for autotiling).
func (level *Level) AreNeighborsTheSame(t *Tile) (top, bottom, left, right bool) {
	x, y, z := t.Coords()

	if n := level.GetTilePtr(x-1, y, z); n != nil && n.Type == t.Type && n.Variant == t.Variant {
		left = true
	}
	if n := level.GetTilePtr(x+1, y, z); n != nil && n.Type == t.Type && n.Variant == t.Variant {
		right = true
	}
	if n := level.GetTilePtr(x, y-1, z); n != nil && n.Type == t.Type && n.Variant == t.Variant {
		top = true
	}
	if n := level.GetTilePtr(x, y+1, z); n != nil && n.Type == t.Type && n.Variant == t.Variant {
		bottom = true
	}
	return
}

// ResolveVariant returns the correct TileVariant for the given tile based on
// its TileDefinition's AutoTile mode and its neighbors.
func (level *Level) ResolveVariant(t *Tile) TileVariant {
	def := TileDefinitions[t.Type]

	switch def.AutoTile {
	case AutoTileWall:
		// 2-variant wall: bottom neighbor connected → Variants[0], else Variants[1]
		_, bottom, _, _ := level.AreNeighborsTheSame(t)
		if bottom {
			return def.Variants[0]
		}
		return def.Variants[1]

	case AutoTileBitmask:
		// 4-bit cardinal bitmask: top=1, bottom=2, left=4, right=8 → 16 variants
		top, bottom, left, right := level.AreNeighborsTheSame(t)
		idx := 0
		if top {
			idx |= 1
		}
		if bottom {
			idx |= 2
		}
		if left {
			idx |= 4
		}
		if right {
			idx |= 8
		}
		if idx < len(def.Variants) {
			return def.Variants[idx]
		}
		return def.Variants[0]

	default:
		// AutoTileNone: use tile.Variant directly
		if t.Variant >= 0 && t.Variant < len(def.Variants) {
			return def.Variants[t.Variant]
		}
		return def.Variants[0]
	}
}

// ─── Index math ──────────────────────────────────────────────────────

func (level *Level) index(x, y, z int) int {
	return x + y*level.Width + z*level.Width*level.Height
}

// ─── Fog of war ──────────────────────────────────────────────────────

// GetSeen reports whether the tile at (x,y,z) has ever been seen.
func (level *Level) GetSeen(x, y, z int) bool {
	if !level.InBounds(x, y, z) {
		return false
	}
	return level.Seen[level.index(x, y, z)]
}

// SetSeen marks the tile at (x,y,z) as seen or unseen.
func (level *Level) SetSeen(x, y, z int, val bool) {
	if !level.InBounds(x, y, z) {
		return
	}
	level.Seen[level.index(x, y, z)] = val
}

// ClearSeen resets all explored state (e.g. when loading a new level).
func (level *Level) ClearSeen() {
	for i := range level.Seen {
		level.Seen[i] = false
	}
}

// ─── Tile access ─────────────────────────────────────────────────────

// GetTilePtr returns a direct *Tile pointer for internal/embedding use.
// Returns nil if out of bounds.
func (level *Level) GetTilePtr(x, y, z int) *Tile {
	if !level.InBounds(x, y, z) {
		return nil
	}
	return &level.Data[level.index(x, y, z)]
}

// GetTilePtrIndex returns a direct *Tile pointer by flat index.
func (level *Level) GetTilePtrIndex(idx int) *Tile {
	if idx < 0 || idx >= len(level.Data) {
		return nil
	}
	return &level.Data[idx]
}

// GetTileAt satisfies LevelInterface.
func (level *Level) GetTileAt(x, y, z int) TileInterface {
	t := level.GetTilePtr(x, y, z)
	if t == nil {
		return nil
	}
	return t
}

// GetTileIndex satisfies LevelInterface.
func (level *Level) GetTileIndex(index int) TileInterface {
	t := level.GetTilePtrIndex(index)
	if t == nil {
		return nil
	}
	return t
}

// UpdateTileAt sets the tile type and variant at (x,y,z) and marks the column dirty.
func (level *Level) UpdateTileAt(x, y, z int, tileType string, variant int) TileInterface {
	if !level.InBounds(x, y, z) {
		return nil
	}
	idx := level.index(x, y, z)
	level.Data[idx].Type = TileNameToIndex[tileType]
	level.Data[idx].Variant = variant
	level.InvalidateSunColumn(x, y)
	return &level.Data[idx]
}

// SetTileType is a convenience method for setting the tile type at Z=0.
func (level *Level) SetTileType(x, y int, t string) error {
	tile := level.GetTilePtr(x, y, 0)
	if tile == nil {
		return errors.New("invalid tile")
	}
	tile.Type = TileNameToIndex[t]
	level.InvalidateSunColumn(x, y)
	return nil
}

// ─── Pathfinding (implements path.Graph) ─────────────────────────────

// PathNeighborIDs appends the flat tile indices of walkable neighbors of tileIdx
// to buf and returns it. Direct array access — no map lookup, no allocation.
func (level *Level) PathNeighborIDs(tileIdx int, buf []int) []int {
	t := &level.Data[tileIdx]
	x, y, z := t.Coords()
	for i := range pathOffsets {
		offset := &pathOffsets[i]
		n := level.GetTilePtr(x+offset[0], y+offset[1], z+offset[2])
		if n == nil {
			continue
		}
		if offset[2] != 0 && !(TileDefinitions[n.Type].StairsUp || TileDefinitions[n.Type].StairsDown) {
			continue
		}
		buf = append(buf, n.Idx)
	}
	return buf
}

// PathCost returns the movement cost between two adjacent tile indices.
// Uses the level's PathCostFunc if set, otherwise DefaultPathCost.
func (level *Level) PathCost(fromIdx, toIdx int) float64 {
	from := &level.Data[fromIdx]
	to := &level.Data[toIdx]
	if level.PathCostFunc != nil {
		return level.PathCostFunc(from, to)
	}
	return DefaultPathCost(from, to)
}

// PathEstimate returns the heuristic estimate (squared Euclidean distance)
// between two tile indices.
func (level *Level) PathEstimate(fromIdx, toIdx int) float64 {
	t1 := &level.Data[fromIdx]
	t2 := &level.Data[toIdx]
	x1, y1, z1 := t1.Coords()
	x2, y2, z2 := t2.Coords()
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return float64(dx*dx + dy*dy + dz*dz)
}

// SizedGraph wraps a Level to validate the full entity footprint at every
// candidate position during A* pathfinding. Use this for entities with a
// SizeComponent so the path only follows tiles where the whole footprint fits
// without crossing solid tiles.
//
// Footprint centering: startX = cx - Width/2, startY = cy - Height/2
// (integer division, matching entityFootprint in PlaceEntity/RemoveEntity).
type SizedGraph struct {
	Level  *Level
	Width  int
	Height int
	Entity *ecs.Entity // the entity being pathfinded — excluded from blocker cost checks
}

// PathNeighborIDs includes a neighbor only when the entity's full footprint
// centered on that neighbor tile is free of solid/out-of-bounds tiles.
func (g *SizedGraph) PathNeighborIDs(tileIdx int, buf []int) []int {
	t := &g.Level.Data[tileIdx]
	x, y, z := t.Coords()
	for i := range pathOffsets {
		offset := &pathOffsets[i]
		nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
		n := g.Level.GetTilePtr(nx, ny, nz)
		if n == nil {
			continue
		}
		if offset[2] != 0 && !(TileDefinitions[n.Type].StairsUp || TileDefinitions[n.Type].StairsDown) {
			continue
		}
		if !g.footprintClear(nx, ny, nz) {
			continue
		}
		buf = append(buf, n.Idx)
	}
	return buf
}

func (g *SizedGraph) footprintClear(cx, cy, z int) bool {
	startX := cx - g.Width/2
	startY := cy - g.Height/2
	for dx := 0; dx < g.Width; dx++ {
		for dy := 0; dy < g.Height; dy++ {
			tile := g.Level.GetTilePtr(startX+dx, startY+dy, z)
			if tile == nil || TileDefinitions[tile.Type].Solid {
				return false
			}
		}
	}
	return true
}

func (g *SizedGraph) PathCost(fromIdx, toIdx int) float64 {
	cost := g.Level.PathCost(fromIdx, toIdx)
	// If the level's cost function penalised this tile due to a solid entity,
	// check whether that entity is actually ourselves.  A sized entity is
	// registered in all its footprint tiles, so the level's PathCostFunc sees
	// it as a blocker at neighbours it already occupies.
	if cost >= 100 && g.Entity != nil {
		to := &g.Level.Data[toIdx]
		if !TileDefinitions[to.Type].Solid {
			x, y, z := to.Coords()
			if blocker := g.Level.GetSolidEntityAt(x, y, z); blocker == g.Entity {
				return 10
			}
		}
	}
	return cost
}

func (g *SizedGraph) PathEstimate(fromIdx, toIdx int) float64 {
	return g.Level.PathEstimate(fromIdx, toIdx)
}

// ─── Time & lighting ─────────────────────────────────────────────────

var sunIntensityTable = [24]int{
	//  0   1   2   3   4   5   6   7   8   9  10  11  12  13  14  15  16  17  18  19  20  21  22  23
	0, 0, 0, 0, 0, 0, 0, 30, 70, 100, 100, 100, 100, 100, 100, 100, 70, 30, 0, 0, 0, 0, 0, 0,
}

func (level *Level) SunIntensity() int {
	return sunIntensityTable[level.Hour]
}

func (level *Level) IsNight() bool {
	return level.Hour < 6 || level.Hour >= 18
}

func (level *Level) NextHour() {
	level.Hour++
	if level.Hour >= 24 {
		level.Hour = 0
		level.Day++
	}
}

// IsTileExposedToSun checks whether the tile at (x,y,z) has clear sky above it.
func (level *Level) IsTileExposedToSun(x, y, z int) bool {
	if !level.InBounds(x, y, z) {
		return false
	}
	for zOffset := z + 1; zOffset < level.Depth; zOffset++ {
		above := level.GetTilePtr(x, y, zOffset)
		if above != nil && !TileDefinitions[above.Type].Air && !TileDefinitions[above.Type].Water {
			return false
		}
	}
	return true
}

// InvalidateSunColumn marks a (x,y) column for incremental lighting update.
func (level *Level) InvalidateSunColumn(x, y int) {
	level.DirtyColumns = append(level.DirtyColumns, y*level.Width+x)
}

// ─── Entity management ───────────────────────────────────────────────

func (level *Level) GetEntities() []*ecs.Entity       { return level.Entities }
func (level *Level) GetStaticEntities() []*ecs.Entity { return level.StaticEntities }

// entityFootprint returns all tile positions occupied by an entity centered at (x,y,z).
// Entities without a SizeComponent are treated as 1×1.
// Centering: startX = x - w/2, startY = y - h/2 (integer division).
func (level *Level) entityFootprint(x, y, z int, entity *ecs.Entity) [][3]int {
	w, h := 1, 1
	if entity.HasComponent(rlcomponents.Size) {
		sc := entity.GetComponent(rlcomponents.Size).(*rlcomponents.SizeComponent)
		if sc.Width > 0 {
			w = sc.Width
		}
		if sc.Height > 0 {
			h = sc.Height
		}
	}
	startX := x - w/2
	startY := y - h/2
	tiles := make([][3]int, 0, w*h)
	for dx := 0; dx < w; dx++ {
		for dy := 0; dy < h; dy++ {
			tx, ty := startX+dx, startY+dy
			if level.InBounds(tx, ty, z) {
				tiles = append(tiles, [3]int{tx, ty, z})
			}
		}
	}
	return tiles
}

// removeFromEntityPos removes a single entity from the entityPos bucket at (x,y,z).
func (level *Level) removeFromEntityPos(x, y, z int, entity *ecs.Entity) {
	if !level.InBounds(x, y, z) {
		return
	}
	key := level.index(x, y, z)
	entities := level.entityPos[key]
	for i := 0; i < len(entities); i++ {
		if entities[i] == entity {
			copy(entities[i:], entities[i+1:])
			entities[len(entities)-1] = nil
			entities = entities[:len(entities)-1]
			if len(entities) == 0 {
				delete(level.entityPos, key)
			} else {
				level.entityPos[key] = entities
			}
			break
		}
	}
}

func (level *Level) PlaceEntity(x, y, z int, entity *ecs.Entity) {
	if !level.InBounds(x, y, z) {
		return
	}

	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

	// Remove from all tiles at old position
	for _, t := range level.entityFootprint(pc.GetX(), pc.GetY(), pc.GetZ(), entity) {
		level.removeFromEntityPos(t[0], t[1], t[2], entity)
	}

	pc.SetPosition(x, y, z)

	// Add to all tiles at new position
	for _, t := range level.entityFootprint(x, y, z, entity) {
		key := level.index(t[0], t[1], t[2])
		level.entityPos[key] = append(level.entityPos[key], entity)
	}
}

func (level *Level) AddEntity(entity *ecs.Entity) {
	if !entity.HasComponent(rlcomponents.Inanimate) {
		level.Entities = append(level.Entities, entity)
	} else {
		level.StaticEntities = append(level.StaticEntities, entity)
	}

	if entity.HasComponent(rlcomponents.Position) {
		pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		level.PlaceEntity(pc.GetX(), pc.GetY(), pc.GetZ(), entity)
	}
}

func (level *Level) RemoveEntity(entity *ecs.Entity) {
	if entity.HasComponent(rlcomponents.Position) {
		pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		x, y, z := pc.GetX(), pc.GetY(), pc.GetZ()
		for _, t := range level.entityFootprint(x, y, z, entity) {
			level.removeFromEntityPos(t[0], t[1], t[2], entity)
		}
	}

	for i := 0; i < len(level.Entities); i++ {
		if level.Entities[i] == entity {
			copy(level.Entities[i:], level.Entities[i+1:])
			level.Entities[len(level.Entities)-1] = nil
			level.Entities = level.Entities[:len(level.Entities)-1]
			return
		}
	}

	for i := 0; i < len(level.StaticEntities); i++ {
		if level.StaticEntities[i] == entity {
			copy(level.StaticEntities[i:], level.StaticEntities[i+1:])
			level.StaticEntities[len(level.StaticEntities)-1] = nil
			level.StaticEntities = level.StaticEntities[:len(level.StaticEntities)-1]
			return
		}
	}
}

func (level *Level) GetEntityAt(x, y, z int) *ecs.Entity {
	if level.InBounds(x, y, z) {
		key := level.index(x, y, z)
		if len(level.entityPos[key]) > 0 {
			return level.entityPos[key][0]
		}
	}
	return nil
}

func (level *Level) GetEntitiesAt(x, y, z int, buffer *[]*ecs.Entity) {
	if level.InBounds(x, y, z) {
		key := level.index(x, y, z)
		if len(level.entityPos[key]) > 0 {
			*buffer = append(*buffer, level.entityPos[key]...)
		}
	} else {
		*buffer = (*buffer)[:0]
	}
}

func (level *Level) GetSolidEntityAt(x, y, z int) *ecs.Entity {
	if level.InBounds(x, y, z) {
		key := level.index(x, y, z)
		for _, entity := range level.entityPos[key] {
			if entity.HasComponent(rlcomponents.Solid) {
				return entity
			}
		}
	}
	return nil
}

func (level *Level) GetEntitiesAround(x, y, z, width, height int, buffer *[]*ecs.Entity) {
	left := max(0, x-width/2)
	right := min(level.Width, x+width/2)
	up := max(0, y-height/2)
	down := min(level.Height, y+height/2)

	*buffer = (*buffer)[:0]

	estimated := (right - left) * (down - up)
	if cap(*buffer) < estimated {
		*buffer = make([]*ecs.Entity, 0, estimated)
	}

	for ix := left; ix < right; ix++ {
		for iy := up; iy < down; iy++ {
			key := level.index(ix, iy, z)
			entities := level.entityPos[key]
			if len(entities) > 0 {
				*buffer = append(*buffer, entities...)
			}
		}
	}
}

func (level *Level) GetClosestEntity(x, y, z, width, height int) *ecs.Entity {
	left := max(0, x-width/2)
	right := min(level.Width, x+width/2)
	up := max(0, y-height/2)
	down := min(level.Height, y+height/2)

	var closest *ecs.Entity
	minDistSq := int(^uint(0) >> 1)

	for ix := left; ix < right; ix++ {
		for iy := up; iy < down; iy++ {
			key := level.index(ix, iy, z)
			for _, entity := range level.entityPos[key] {
				pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
				dx := pc.GetX() - x
				dy := pc.GetY() - y
				distSq := dx*dx + dy*dy
				if distSq < minDistSq {
					minDistSq = distSq
					closest = entity
				}
			}
		}
	}
	return closest
}

func (level *Level) GetClosestEntityMatching(x, y, z, width, height int, exclude *ecs.Entity, match func(*ecs.Entity) bool) *ecs.Entity {
	left := max(0, x-width/2)
	right := min(level.Width, x+width/2)
	up := max(0, y-height/2)
	down := min(level.Height, y+height/2)

	var closest *ecs.Entity
	minDistSq := int(^uint(0) >> 1)

	cx := x
	cy := y

	maxRadius := max(right-left, down-up) / 2
	for radius := 0; radius <= maxRadius; radius++ {
		for dx := -radius; dx <= radius; dx++ {
			for dy := -radius; dy <= radius; dy++ {
				if utility.Abs(dx) != radius && utility.Abs(dy) != radius {
					continue
				}
				ix := cx + dx
				iy := cy + dy
				if ix < left || ix >= right || iy < up || iy >= down {
					continue
				}
				key := level.index(ix, iy, z)
				for _, entity := range level.entityPos[key] {
					if entity == exclude {
						continue
					}
					if match(entity) {
						pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
						dx2 := pc.GetX() - x
						dy2 := pc.GetY() - y
						distSq := dx2*dx2 + dy2*dy2
						if distSq < minDistSq {
							minDistSq = distSq
							closest = entity
						}
					}
				}
			}
		}
		if closest != nil {
			break
		}
	}
	return closest
}
