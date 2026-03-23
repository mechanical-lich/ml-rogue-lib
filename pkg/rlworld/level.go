package rlworld

import "github.com/mechanical-lich/mlge/ecs"

type LevelInterface interface {
	// Dimensions
	GetWidth() int
	GetHeight() int
	GetDepth() int

	// Bounds
	InBounds(x, y, z int) bool

	// Tile access
	GetTileAt(x, y, z int) TileInterface
	GetTileIndex(index int) TileInterface
	UpdateTileAt(x, y, z int, tileType string, variant int) TileInterface
	SetTileType(x, y int, t string) error

	// Time & lighting
	SunIntensity() int
	IsNight() bool
	IsTileExposedToSun(x, y, z int) bool
	InvalidateSunColumn(x, y int)
	NextHour()

	// Entity management
	PlaceEntity(x, y, z int, entity *ecs.Entity)
	AddEntity(entity *ecs.Entity)
	RemoveEntity(entity *ecs.Entity)
	GetEntityAt(x, y, z int) *ecs.Entity
	GetEntitiesAt(x, y, z int, buffer *[]*ecs.Entity)
	GetEntitiesAround(x, y, z, width, height int, buffer *[]*ecs.Entity)
	GetClosestEntity(x, y, z, width, height int) *ecs.Entity
	GetSolidEntityAt(x, y, z int) *ecs.Entity
	GetSolidEntitiesAt(x, y, z int, buf *[]*ecs.Entity)
	GetClosestEntityMatching(x, y, z, width, height int, exclude *ecs.Entity, match func(*ecs.Entity) bool) *ecs.Entity

	// Entity lists (read-only iteration)
	GetEntities() []*ecs.Entity
	GetStaticEntities() []*ecs.Entity
	AreNeighborsTheSame(t *Tile) (top, bottom, left, right bool)
}
