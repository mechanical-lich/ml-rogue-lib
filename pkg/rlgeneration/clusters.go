package rlgeneration

import (
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/utility"
)

// EntityFactory is a function that creates a named entity at the given position.
// Implement this using your game's factory/blueprint system.
type EntityFactory func(name string, x, y, z int) (*ecs.Entity, error)

// CreateClusterOfTiles places tiles of the given type in a random-walk cluster
// starting at (x,y,z) with the given size.
func CreateClusterOfTiles(level rlworld.LevelInterface, x, y, z, size int, tileName string) {
	for i := 0; i < size; i++ {
		if utility.GetRandom(1, 6) == 1 {
			x++
		}
		if utility.GetRandom(1, 6) == 1 {
			x--
		}
		if utility.GetRandom(1, 6) == 1 {
			y++
		}
		if utility.GetRandom(1, 6) == 1 {
			y--
		}
		if level.InBounds(x, y, z) {
			level.SetTileType(x, y, tileName)
		}
	}
}

// Create3DClusterOfTiles is the same as CreateClusterOfTiles but also walks along Z.
func Create3DClusterOfTiles(level rlworld.LevelInterface, x, y, z, size int, tileName string) {
	for i := 0; i < size; i++ {
		if utility.GetRandom(1, 6) == 1 {
			x++
		}
		if utility.GetRandom(1, 6) == 1 {
			x--
		}
		if utility.GetRandom(1, 6) == 1 {
			y++
		}
		if utility.GetRandom(1, 6) == 1 {
			y--
		}
		if utility.GetRandom(1, 6) == 1 {
			z++
		}
		if utility.GetRandom(1, 6) == 1 {
			z--
		}
		if level.InBounds(x, y, z) {
			level.SetTileType(x, y, tileName)
		}
	}
}

// CreateClusterOfEntitiesTagged is like CreateClusterOfEntities but calls onSpawn
// for each successfully placed entity. Use this to attach game-specific components
// (e.g., a settlement ownership tag) to every spawned entity.
func CreateClusterOfEntitiesTagged(
	level rlworld.LevelInterface,
	x, y, z, size int,
	entityName string,
	factory EntityFactory,
	maxRetries int,
	onSpawn func(*ecs.Entity),
) {
	for i := 0; i < size; {
		if utility.GetRandom(1, 6) == 1 {
			x++
		}
		if utility.GetRandom(1, 6) == 1 {
			x--
		}
		if utility.GetRandom(1, 6) == 1 {
			y++
		}
		if utility.GetRandom(1, 6) == 1 {
			y--
		}

		if !level.InBounds(x, y, z) {
			i++
			continue
		}

		tile := level.GetTileAt(x, y, z)
		if tile == nil || tile.IsSolid() || tile.IsWater() || tile.IsAir() || level.GetEntityAt(x, y, z) != nil {
			if maxRetries > 0 {
				maxRetries--
				continue
			}
			i++
			continue
		}

		entity, err := factory(entityName, x, y, z)
		if err == nil && entity != nil {
			if onSpawn != nil {
				onSpawn(entity)
			}
			level.AddEntity(entity)
		}
		i++
	}
}

// CreateClusterOfEntities spawns entities via factory in a random-walk cluster.
// Skips positions that are solid, water, air, or already occupied.
// maxRetries controls how many times to retry a blocked position before moving on.
func CreateClusterOfEntities(
	level rlworld.LevelInterface,
	x, y, z, size int,
	entityName string,
	factory EntityFactory,
	maxRetries int,
) {
	for i := 0; i < size; {
		if utility.GetRandom(1, 6) == 1 {
			x++
		}
		if utility.GetRandom(1, 6) == 1 {
			x--
		}
		if utility.GetRandom(1, 6) == 1 {
			y++
		}
		if utility.GetRandom(1, 6) == 1 {
			y--
		}

		if !level.InBounds(x, y, z) {
			i++
			continue
		}

		tile := level.GetTileAt(x, y, z)
		if tile == nil || tile.IsSolid() || tile.IsWater() || tile.IsAir() || level.GetEntityAt(x, y, z) != nil {
			if maxRetries > 0 {
				maxRetries--
				continue
			}
			i++
			continue
		}

		entity, err := factory(entityName, x, y, z)
		if err == nil && entity != nil {
			level.AddEntity(entity)
		}
		i++
	}
}
