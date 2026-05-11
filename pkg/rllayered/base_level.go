package rllayered

import (
	"errors"
	"log"
	"runtime"
	"sync"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/utility"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
)

// Level is a GC-optimized 3D tile container with spatial entity indexing.
// Each cell carries Floor/Middle/Ceiling slots; see Tile.
type Level struct {
	Data           []Tile
	Seen           []bool
	Entities       []*ecs.Entity
	StaticEntities []*ecs.Entity
	entityPos      map[int][]*ecs.Entity
	Width          int
	Height         int
	Depth          int
	Hour           int
	Day            int

	DirtyColumns []int

	// PathCostFunc is called by PathCost. If nil, DefaultPathCost is used.
	PathCostFunc func(from, to *Tile) float64
}

var _ rlworld.LevelInterface = (*Level)(nil)

// NewLevel creates a Level with the given dimensions. Tiles start with empty
// slots (every Slot.Type == 0). Callers should paint terrain afterwards.
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

// InitTiles initializes index/width/height bookkeeping on every tile. Slots
// stay empty — primers paint terrain by name.
func (level *Level) InitTiles() {
	log.Println("Initializing layered tiles")
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

// AreNeighborsTheSame compares the Middle slot of the four cardinal neighbors
// against t.Middle (autotiling is a Middle-layer concern; floors don't
// autotile in the POC).
func (level *Level) AreNeighborsTheSame(t *Tile) (top, bottom, left, right bool) {
	top, bottom, left, right, _, _, _, _ = level.areNeighborsTheSame8(t)
	return
}

// areNeighborsTheSame8 returns matches in all 8 directions. Used by the
// autotile resolver to detect "fully surrounded" cells and pick the interior
// variant (16) when one is provided.
func (level *Level) areNeighborsTheSame8(t *Tile) (top, bottom, left, right, ne, nw, se, sw bool) {
	x, y, z := t.Coords()
	if t.Middle.IsEmpty() {
		return
	}
	def := TileDefinitions[t.Middle.Type]
	isAutoTile := def.AutoTile != AutoTileNone

	sameType := func(n *Tile) bool {
		if n == nil || n.Middle.IsEmpty() {
			return false
		}
		if isAutoTile {
			return n.Middle.Type == t.Middle.Type
		}
		return n.Middle.Type == t.Middle.Type && n.Middle.Variant == t.Middle.Variant
	}

	if sameType(level.GetTilePtr(x-1, y, z)) {
		left = true
	}
	if sameType(level.GetTilePtr(x+1, y, z)) {
		right = true
	}
	if sameType(level.GetTilePtr(x, y-1, z)) {
		top = true
	}
	if sameType(level.GetTilePtr(x, y+1, z)) {
		bottom = true
	}
	if sameType(level.GetTilePtr(x+1, y-1, z)) {
		ne = true
	}
	if sameType(level.GetTilePtr(x-1, y-1, z)) {
		nw = true
	}
	if sameType(level.GetTilePtr(x+1, y+1, z)) {
		se = true
	}
	if sameType(level.GetTilePtr(x-1, y+1, z)) {
		sw = true
	}
	return
}

// ResolveVariant returns the TileVariant to render for the Middle slot of t
// based on its TileDefinition's AutoTile mode and neighbors.
func (level *Level) ResolveVariant(t *Tile) TileVariant {
	if t.Middle.IsEmpty() {
		return TileVariant{}
	}
	def := TileDefinitions[t.Middle.Type]

	switch def.AutoTile {
	case AutoTileWall:
		_, bottom, _, _ := level.AreNeighborsTheSame(t)
		want := 1
		if bottom {
			want = 0
		}
		for i := range def.Variants {
			if def.Variants[i].Variant == want {
				return def.Variants[i]
			}
		}
		return def.Variants[0]

	case AutoTileBitmask:
		top, bottom, left, right, ne, nw, se, sw := level.areNeighborsTheSame8(t)
		// "Fully interior": all 8 neighbors match. If the artist provided a
		// variant 16 sprite, use it — that's the no-borders-anywhere fill.
		if top && bottom && left && right && ne && nw && se && sw {
			for i := range def.Variants {
				if def.Variants[i].Variant == 16 {
					return def.Variants[i]
				}
			}
		}
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
		for i := range def.Variants {
			if def.Variants[i].Variant == idx {
				return def.Variants[i]
			}
		}
		return def.Variants[0]

	case AutoTileBlob47:
		top, bottom, left, right, ne, nw, se, sw := level.areNeighborsTheSame8(t)
		mask := uint8(0)
		if top {
			mask |= BlobBitN
		}
		if bottom {
			mask |= BlobBitS
		}
		if left {
			mask |= BlobBitW
		}
		if right {
			mask |= BlobBitE
		}
		if ne {
			mask |= BlobBitNE
		}
		if nw {
			mask |= BlobBitNW
		}
		if se {
			mask |= BlobBitSE
		}
		if sw {
			mask |= BlobBitSW
		}
		mask = PruneBlobMask(mask)
		// Variants for AutoTileBlob47 use the pruned mask value (0..255) as
		// their Variant field — no ordering convention. The artist labels
		// each sheet cell with its mask and lays it out however they like.
		for i := range def.Variants {
			if def.Variants[i].Variant == int(mask) {
				return def.Variants[i]
			}
		}
		return def.Variants[0]

	default:
		if t.Middle.Variant >= 0 && t.Middle.Variant < len(def.Variants) {
			return def.Variants[t.Middle.Variant]
		}
		return def.Variants[0]
	}
}

// ResolveVariantForSlot returns the TileVariant for a specific slot of a tile.
// Used by renderers that draw Floor and Ceiling layers (which don't autotile
// against Middle neighbors).
func (level *Level) ResolveVariantForSlot(s Slot) TileVariant {
	if s.IsEmpty() {
		return TileVariant{}
	}
	def := TileDefinitions[s.Type]
	if s.Variant >= 0 && s.Variant < len(def.Variants) {
		return def.Variants[s.Variant]
	}
	if len(def.Variants) > 0 {
		return def.Variants[0]
	}
	return TileVariant{}
}

// ─── Index math ──────────────────────────────────────────────────────

func (level *Level) index(x, y, z int) int {
	return x + y*level.Width + z*level.Width*level.Height
}

// ─── Fog of war ──────────────────────────────────────────────────────

func (level *Level) GetSeen(x, y, z int) bool {
	if !level.InBounds(x, y, z) {
		return false
	}
	return level.Seen[level.index(x, y, z)]
}

func (level *Level) SetSeen(x, y, z int, val bool) {
	if !level.InBounds(x, y, z) {
		return
	}
	level.Seen[level.index(x, y, z)] = val
}

func (level *Level) ClearSeen() {
	for i := range level.Seen {
		level.Seen[i] = false
	}
}

// ─── Tile access ─────────────────────────────────────────────────────

func (level *Level) GetTilePtr(x, y, z int) *Tile {
	if !level.InBounds(x, y, z) {
		return nil
	}
	return &level.Data[level.index(x, y, z)]
}

func (level *Level) GetTilePtrIndex(idx int) *Tile {
	if idx < 0 || idx >= len(level.Data) {
		return nil
	}
	return &level.Data[idx]
}

func (level *Level) GetTileAt(x, y, z int) rlworld.TileInterface {
	t := level.GetTilePtr(x, y, z)
	if t == nil {
		return nil
	}
	return t
}

func (level *Level) GetTileIndex(index int) rlworld.TileInterface {
	t := level.GetTilePtrIndex(index)
	if t == nil {
		return nil
	}
	return t
}

// UpdateTileAt satisfies rlworld.LevelInterface by routing to PaintTile.
// External callers that don't know about layers continue to work — the tile
// lands in whichever slot its TileDefinition declares.
func (level *Level) UpdateTileAt(x, y, z int, tileType string, variant int) rlworld.TileInterface {
	return level.PaintTile(x, y, z, tileType, variant)
}

// PaintTile places `tileType` into the slot indicated by its TileDefinition's
// Layer field. This is the layered-cell replacement for rlworld.UpdateTileAt.
func (level *Level) PaintTile(x, y, z int, tileType string, variant int) rlworld.TileInterface {
	if !level.InBounds(x, y, z) {
		return nil
	}
	idx := level.index(x, y, z)
	t := &level.Data[idx]
	typeIdx, ok := TileNameToIndex[tileType]
	if !ok {
		return t
	}
	def := &TileDefinitions[typeIdx]
	slot := Slot{Type: typeIdx, Variant: variant}
	switch def.LayerOf() {
	case LayerFloor:
		t.Floor = slot
	case LayerCeiling:
		t.Ceiling = slot
	default:
		t.Middle = slot
	}
	level.InvalidateSunColumn(x, y)
	return t
}

// SetFloor / SetMiddle / SetCeiling are direct slot setters for cases where
// the caller has decided the layer explicitly (e.g. primers placing terrain
// where a single tile name has both floor and middle semantics).
func (level *Level) SetFloor(x, y, z int, tileType string, variant int) {
	if !level.InBounds(x, y, z) {
		return
	}
	idx, ok := TileNameToIndex[tileType]
	if !ok {
		return
	}
	level.Data[level.index(x, y, z)].Floor = Slot{Type: idx, Variant: variant}
	level.InvalidateSunColumn(x, y)
}

func (level *Level) SetMiddle(x, y, z int, tileType string, variant int) {
	if !level.InBounds(x, y, z) {
		return
	}
	idx, ok := TileNameToIndex[tileType]
	if !ok {
		return
	}
	level.Data[level.index(x, y, z)].Middle = Slot{Type: idx, Variant: variant}
	level.InvalidateSunColumn(x, y)
}

func (level *Level) SetCeiling(x, y, z int, tileType string, variant int) {
	if !level.InBounds(x, y, z) {
		return
	}
	idx, ok := TileNameToIndex[tileType]
	if !ok {
		return
	}
	level.Data[level.index(x, y, z)].Ceiling = Slot{Type: idx, Variant: variant}
	level.InvalidateSunColumn(x, y)
}

// ClearMiddle removes the Middle slot (used for mining/digging).
func (level *Level) ClearMiddle(x, y, z int) {
	if !level.InBounds(x, y, z) {
		return
	}
	level.Data[level.index(x, y, z)].Middle = Slot{}
	level.InvalidateSunColumn(x, y)
}

// ClearFloor removes the Floor slot (used for digging through ground).
func (level *Level) ClearFloor(x, y, z int) {
	if !level.InBounds(x, y, z) {
		return
	}
	level.Data[level.index(x, y, z)].Floor = Slot{}
	level.InvalidateSunColumn(x, y)
}

// SetTileType is a legacy convenience that paints into the slot dictated by
// the tile definition's Layer at z=0.
func (level *Level) SetTileType(x, y int, t string) error {
	if !level.InBounds(x, y, 0) {
		return errors.New("invalid tile")
	}
	level.PaintTile(x, y, 0, t, 0)
	return nil
}

// IsWalkable reports whether an entity can occupy (x,y,z). A walkable cell
// needs ground (Floor non-empty) and a non-blocking Middle.
func (level *Level) IsWalkable(x, y, z int) bool {
	if !level.InBounds(x, y, z) {
		return false
	}
	t := &level.Data[level.index(x, y, z)]
	if t.Floor.IsEmpty() {
		return false
	}
	if !t.Middle.IsEmpty() {
		mid := TileDefinitions[t.Middle.Type]
		if mid.Solid || mid.Water {
			return false
		}
	}
	return true
}

// ─── Pathfinding (implements path.Graph) ─────────────────────────────

func (level *Level) PathNeighborIDs(tileIdx int, buf []int) []int {
	t := &level.Data[tileIdx]
	x, y, z := t.Coords()
	for i := range pathOffsets {
		offset := &pathOffsets[i]
		n := level.GetTilePtr(x+offset[0], y+offset[1], z+offset[2])
		if n == nil {
			continue
		}
		if offset[2] != 0 {
			if n.Middle.IsEmpty() {
				continue
			}
			def := TileDefinitions[n.Middle.Type]
			if !(def.StairsUp || def.StairsDown) {
				continue
			}
		}
		buf = append(buf, n.Idx)
	}
	return buf
}

func (level *Level) PathCost(fromIdx, toIdx int) float64 {
	from := &level.Data[fromIdx]
	to := &level.Data[toIdx]
	if level.PathCostFunc != nil {
		return level.PathCostFunc(from, to)
	}
	return DefaultPathCost(from, to)
}

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

// SizedGraph mirrors rlworld.SizedGraph for multi-tile entity footprints.
type SizedGraph struct {
	Level  *Level
	Width  int
	Height int
	Entity *ecs.Entity
}

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
		if offset[2] != 0 {
			if n.Middle.IsEmpty() {
				continue
			}
			def := TileDefinitions[n.Middle.Type]
			if !(def.StairsUp || def.StairsDown) {
				continue
			}
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
			if tile == nil {
				return false
			}
			if !tile.Middle.IsEmpty() && TileDefinitions[tile.Middle.Type].Solid {
				return false
			}
		}
	}
	return true
}

func (g *SizedGraph) PathCost(fromIdx, toIdx int) float64 {
	cost := g.Level.PathCost(fromIdx, toIdx)
	if cost >= 100 && g.Entity != nil {
		to := &g.Level.Data[toIdx]
		isSolid := !to.Middle.IsEmpty() && TileDefinitions[to.Middle.Type].Solid
		if !isSolid {
			x, y, z := to.Coords()
			var solidBuf []*ecs.Entity
			g.Level.GetSolidEntitiesAt(x, y, z, &solidBuf)
			onlySelf := len(solidBuf) > 0
			for _, e := range solidBuf {
				if e != g.Entity {
					onlySelf = false
					break
				}
			}
			if onlySelf {
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

// IsTileExposedToSun walks upward from (x,y,z) and returns true if every cell
// above has an empty/air/water Middle (sky-transparent).
func (level *Level) IsTileExposedToSun(x, y, z int) bool {
	if !level.InBounds(x, y, z) {
		return false
	}
	for zOffset := z + 1; zOffset < level.Depth; zOffset++ {
		above := level.GetTilePtr(x, y, zOffset)
		if above == nil {
			continue
		}
		if above.Middle.IsEmpty() {
			continue
		}
		def := TileDefinitions[above.Middle.Type]
		if !def.Air && !def.Water {
			return false
		}
	}
	return true
}

func (level *Level) InvalidateSunColumn(x, y int) {
	level.DirtyColumns = append(level.DirtyColumns, y*level.Width+x)
}

// ─── Entity management ───────────────────────────────────────────────

func (level *Level) GetEntities() []*ecs.Entity       { return level.Entities }
func (level *Level) GetStaticEntities() []*ecs.Entity { return level.StaticEntities }

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
	for _, t := range level.entityFootprint(pc.GetX(), pc.GetY(), pc.GetZ(), entity) {
		level.removeFromEntityPos(t[0], t[1], t[2], entity)
	}
	pc.SetPosition(x, y, z)
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

func (level *Level) GetSolidEntitiesAt(x, y, z int, buf *[]*ecs.Entity) {
	if level.InBounds(x, y, z) {
		key := level.index(x, y, z)
		for _, entity := range level.entityPos[key] {
			if entity.HasComponent(rlcomponents.Solid) {
				*buf = append(*buf, entity)
			}
		}
	}
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
