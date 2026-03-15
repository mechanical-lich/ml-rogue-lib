---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

This guide covers installation and a minimal integration example using ML Rogue Lib.

## Prerequisites

- **Go 1.25+**
- A project already using [MLGE](https://mechanical-lich.github.io/mlge) (or at minimum `github.com/mechanical-lich/mlge/ecs`)

## Installation

```bash
go get github.com/mechanical-lich/ml-rogue-lib
```

## Core Concepts

ML Rogue Lib is built around MLGE's ECS. Every game object is an `*ecs.Entity` carrying a set of `rlcomponents` structs. Systems iterate entities each frame and act on whichever components they require.

The library does **not** own your game loop or your concrete level type. Instead it defines two interfaces — `rlworld.LevelInterface` and `rlworld.TileInterface` — that your level implementation must satisfy. All systems and helpers accept these interfaces, keeping the library decoupled from any specific game.

## Minimal Integration

### 1. Implement the world interfaces

```go
// MyTile satisfies rlworld.TileInterface.
type MyTile struct { X, Y, Z int; solid bool }

func (t *MyTile) Coords() (int, int, int)   { return t.X, t.Y, t.Z }
func (t *MyTile) IsSolid() bool              { return t.solid }
func (t *MyTile) IsWater() bool              { return false }
func (t *MyTile) IsAir() bool                { return false }
// ... implement PathID, PathNeighborsAppend, etc.

// MyLevel satisfies rlworld.LevelInterface.
type MyLevel struct { tiles []MyTile; entities []*ecs.Entity }

func (l *MyLevel) GetWidth() int  { return 64 }
func (l *MyLevel) GetHeight() int { return 64 }
// ... implement remaining interface methods
```

### 2. Spawn entities with components

```go
import (
    "github.com/mechanical-lich/mlge/ecs"
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
)

func spawnPlayer(level *MyLevel, x, y int) *ecs.Entity {
    e := &ecs.Entity{Blueprint: "player"}
    e.AddComponent(&rlcomponents.PositionComponent{X: x, Y: y, Z: 0})
    e.AddComponent(&rlcomponents.HealthComponent{MaxHealth: 20, Health: 20})
    e.AddComponent(&rlcomponents.StatsComponent{
        AC: 12, Str: 14, Dex: 12,
        BasicAttackDice: "1d6",
    })
    e.AddComponent(&rlcomponents.InitiativeComponent{DefaultValue: 10, Ticks: 10})
    e.AddComponent(&rlcomponents.DescriptionComponent{Name: "Player"})
    e.AddComponent(&rlcomponents.InventoryComponent{})
    level.AddEntity(e)
    return e
}
```

### 3. Register systems and run the game loop

```go
import (
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
    "github.com/mechanical-lich/mlge/ecs"
)

type Game struct {
    level    *MyLevel
    systemMgr ecs.SystemManager
    cleanup  rlsystems.CleanUpSystem
}

func NewGame() *Game {
    g := &Game{level: newMyLevel()}

    // Register systems.
    g.systemMgr.AddSystem(rlsystems.NewAISystem())
    g.systemMgr.AddSystem(&rlsystems.InitiativeSystem{Speed: 1})
    g.systemMgr.AddSystem(&rlsystems.StatusConditionSystem{})

    // Wire up extension hooks.
    g.cleanup.OnEntityDead = func(level rlworld.LevelInterface, e *ecs.Entity) {
        // spawn loot, award XP, play sounds…
    }
    return g
}

func (g *Game) Update() error {
    // 1. Strip MyTurn and remove dead entities.
    g.cleanup.Update(g.level)

    // 2. Run all registered systems for every entity.
    g.systemMgr.UpdateSystemsForEntities(g.level, g.level.GetEntities())
    return nil
}
```

## Extension Hooks

Every system in `rlsystems` exposes callback fields (e.g. `OnEntityDead`, `OnHostileAttack`, `OnEntityTurn`) rather than hard-coding game-specific behaviour. Assign Go functions to these fields to layer your game's logic on top of the built-in mechanics.

```go
aiSystem := rlsystems.NewAISystem()
aiSystem.OnHostileAttack = func(level rlworld.LevelInterface, attacker, target *ecs.Entity) {
    // play sfx, shake camera, etc.
}
```

See individual package pages for the full list of hooks each system exposes.
