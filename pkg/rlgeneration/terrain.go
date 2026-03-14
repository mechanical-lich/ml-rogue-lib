package rlgeneration

import (
	"log"
	"runtime"
	"sync"
	"time"

	"github.com/aquilax/go-perlin"
	"github.com/mechanical-lich/mg-rogue-lib/pkg/rlworld"
)

const (
	defaultAlpha = 6.
	defaultBeta  = 5.
	defaultN     = 2
)

// OverworldTileFunc is called for every tile during overworld generation.
// It receives the tile coordinates, the Perlin noise value at that point, and the
// startingZ layer. Return the tile name to place at (x,y,z), or "" to leave it as-is.
type OverworldTileFunc func(x, y, z int, noise float64, startingZ int) string

// IslandSurfaceFunc is called for each (x,y) position during the island surface pass.
// noise is the Perlin value and normDist is the normalized distance from center [0,1].
// Return the surface tile name (e.g., "grass", "water", "beach", "mountain").
type IslandSurfaceFunc func(x, y int, noise, normDist float64) string

// IslandFillFunc is called for every tile (x,y,z) during the island fill pass.
// surfaceType is the tile name determined by IslandSurfaceFunc for this (x,y) column.
// noise is the per-voxel Perlin value; normDist is the normalized radial distance.
// Return the tile name to place, or "" to leave it as-is.
type IslandFillFunc func(x, y, z, startingZ int, surfaceType string, noise, normDist float64) string

// PerlinConfig controls the Perlin noise parameters. Use DefaultPerlinConfig for the
// same values as the original fantasy_settlements generators.
type PerlinConfig struct {
	Alpha int64
	Beta  int64
	N     int32
	Seed  int64 // 0 = use current time
}

// DefaultPerlinConfig returns the same Perlin parameters used by fantasy_settlements.
func DefaultPerlinConfig() PerlinConfig {
	return PerlinConfig{Alpha: defaultAlpha, Beta: defaultBeta, N: defaultN}
}

func newPerlin(cfg PerlinConfig) *perlin.Perlin {
	seed := cfg.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return perlin.NewPerlin(float64(cfg.Alpha), float64(cfg.Beta), cfg.N, seed)
}

// GenerateOverworldThreaded fills the level with terrain using Perlin noise,
// dispatching work across all available CPU cores.
//
// tileAt is called for every (x,y,z) to decide which tile to place. The game
// controls all tile-name logic; the library only drives the threading and noise.
//
// Example tileAt for a simple overworld:
//
//	func(x, y, z int, noise float64, startingZ int) string {
//	    if z < startingZ { return "mountain" }
//	    if z == startingZ {
//	        if noise < -0.1 { return "beach" }
//	        return "grass"
//	    }
//	    if noise >= 0.2 { return "mountain" }
//	    return ""
//	}
func GenerateOverworldThreaded(
	level rlworld.LevelInterface,
	startingZ int,
	cfg PerlinConfig,
	tileAt OverworldTileFunc,
) {
	log.Println("rlgeneration: generating overworld terrain")
	p := newPerlin(cfg)

	width := level.GetWidth()
	height := level.GetHeight()
	depth := level.GetDepth()

	const chunkSize = 128
	type chunk struct{ z, xStart, xEnd int }

	var wg sync.WaitGroup
	chunkChan := make(chan chunk, 16)

	worker := func() {
		for c := range chunkChan {
			for x := c.xStart; x < c.xEnd; x++ {
				for y := 0; y < height; y++ {
					noise := p.Noise3D(float64(x)/100, float64(y)/100, float64(c.z)/10)
					if name := tileAt(x, y, c.z, noise, startingZ); name != "" {
						level.UpdateTileAt(x, y, c.z, name, 0)
					}
				}
			}
			wg.Done()
		}
	}

	numWorkers := runtime.NumCPU() - 1
	if numWorkers < 1 {
		numWorkers = 1
	}
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	for z := 0; z < depth; z++ {
		for xStart := 0; xStart < width; xStart += chunkSize {
			xEnd := xStart + chunkSize
			if xEnd > width {
				xEnd = width
			}
			wg.Add(1)
			chunkChan <- chunk{z: z, xStart: xStart, xEnd: xEnd}
		}
	}
	wg.Wait()
	close(chunkChan)
}

// GenerateIslandThreaded fills the level with island terrain using Perlin noise and
// a radial distance falloff, dispatching fill work across all available CPU cores.
//
// The generator runs in two passes:
//  1. Surface pass (single-threaded): calls surfaceAt for every (x,y) to determine
//     the tile type at startingZ. Results are stored and reused in the fill pass.
//  2. Fill pass (multi-threaded): calls fillAt for every (x,y,z) with the surface
//     type for that column, the per-voxel noise, and normDist.
//
// Example surfaceAt for a basic island:
//
//	func(x, y int, noise, normDist float64) string {
//	    if normDist > 0.85 || noise < -0.1 { return "water" }
//	    if normDist > 0.7              { return "beach" }
//	    if noise >= 0.2 && normDist < 0.8 { return "mountain" }
//	    return "grass"
//	}
func GenerateIslandThreaded(
	level rlworld.LevelInterface,
	startingZ int,
	cfg PerlinConfig,
	surfaceAt IslandSurfaceFunc,
	fillAt IslandFillFunc,
) {
	log.Println("rlgeneration: generating island terrain")
	p := newPerlin(cfg)

	width := level.GetWidth()
	height := level.GetHeight()
	depth := level.GetDepth()

	centerX := float64(width) / 2
	centerY := float64(height) / 2
	maxDist := centerX
	if centerY < maxDist {
		maxDist = centerY
	}

	normDist := func(x, y int) float64 {
		dx := float64(x) - centerX
		dy := float64(y) - centerY
		d := (dx*dx + dy*dy) / (maxDist * maxDist)
		if d > 1 {
			d = 1
		}
		return d
	}

	// Surface pass: determine tile type at startingZ for each (x,y).
	surface := make([][]string, width)
	for x := 0; x < width; x++ {
		surface[x] = make([]string, height)
		for y := 0; y < height; y++ {
			nd := normDist(x, y)
			noise := p.Noise3D(float64(x)/100, float64(y)/100, float64(startingZ)/10)
			// Apply radial falloff: bias toward ocean at edges.
			noise = noise*(1-nd) - nd*0.5
			surface[x][y] = surfaceAt(x, y, noise, nd)
		}
	}

	// Fill pass: threaded per-column chunk.
	const chunkSize = 128
	type chunk struct{ xStart, xEnd int }

	var wg sync.WaitGroup
	chunkChan := make(chan chunk, 16)

	worker := func() {
		for c := range chunkChan {
			for x := c.xStart; x < c.xEnd; x++ {
				for y := 0; y < height; y++ {
					for z := 0; z < depth; z++ {
						nd := normDist(x, y)
						noise := p.Noise3D(float64(x)/100, float64(y)/100, float64(z)/10)
						noise = noise*(1-nd) - nd*0.5
						if name := fillAt(x, y, z, startingZ, surface[x][y], noise, nd); name != "" {
							level.UpdateTileAt(x, y, z, name, 0)
						}
					}
				}
			}
			wg.Done()
		}
	}

	numWorkers := runtime.NumCPU() - 1
	if numWorkers < 1 {
		numWorkers = 1
	}
	for i := 0; i < numWorkers; i++ {
		go worker()
	}

	for xStart := 0; xStart < width; xStart += chunkSize {
		xEnd := xStart + chunkSize
		if xEnd > width {
			xEnd = width
		}
		wg.Add(1)
		chunkChan <- chunk{xStart: xStart, xEnd: xEnd}
	}
	wg.Wait()
	close(chunkChan)
}
