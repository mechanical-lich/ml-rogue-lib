---
layout: default
title: Home
nav_order: 1
---

# ML Rogue Lib

Reusable roguelike and tile-based game packages built on top of [MLGE](https://mechanical-lich.github.io/mlge).

## Overview

ML Rogue Lib is a Go library providing the building blocks for roguelike and tile-based games. It supplies ready-made ECS components, AI behaviors, a turn-based combat system, procedural generation helpers, and game systems — all designed to plug into any project that uses [MLGE's](https://mechanical-lich.github.io/mlge) ECS and world interfaces.

Rather than prescribing a specific game loop, the library is intentionally open: systems expose extension hooks so each game can supply its own sounds, effects, UI feedback, and win conditions without forking library code.

## Key Features

- **Rich Component Library** — Position, Health, Stats, Initiative, Inventory, AI, status effects, doors, lighting, and more
- **Turn-Based Initiative** — Tick-down initiative counter with nocturnal/diurnal scheduling and alert overrides
- **D&D-Style Combat** — Full attack-roll pipeline: to-hit vs AC, damage dice, resistances, weaknesses, and status effect transfer
- **AI Behaviours** — Wander, Hostile (pursue + attack), and Defensive AIs with pluggable pathfinding
- **Status Conditions** — Poisoned, Burning, Alerted, Regeneration — each decaying over turns
- **Door System** — Open/close state synced to sprite coordinates via a minimal interface
- **Fog of War** — Bresenham line-of-sight, field-of-view radius sweep, and per-tile explored state built into the base level
- **Level Generation** — Perlin-noise overworld and island generators, room carvers, and entity cluster spawners
- **World Interfaces & Base Types** — `LevelInterface` and `TileInterface` contracts with ready-to-use GC-optimized `Level`, `Tile`, and `TileDefinition` base implementations
- **Entity Helpers** — Stateless movement, facing, swapping, eating, and death detection in `rlentity`
- **AI Navigation** — Target tracking, range checks, and path-following in `rlai`
- **Cleanup System** — Dead-entity removal with drops, XP hooks, and `MyTurn` strip each frame

## Packages

| Package | Description |
|---------|-------------|
| [`path`](path.html) | Graph-centric A\* pathfinding with a GC-invisible hot path |
| [`rlworld`](rlworld.html) | `LevelInterface` / `TileInterface` contracts + GC-optimized base `Level`, `Tile`, and `TileDefinition` types |
| [`rlfov`](rlfov.html) | Line-of-sight (Bresenham) and field-of-view with built-in fog-of-war explored state |
| [`rlcomponents`](rlcomponents.html) | ECS component types and structs |
| [`rlai`](rlai.html) | AI navigation helpers (target tracking, range checks, path following) |
| [`rlcombat`](rlcombat.html) | D&D-style melee combat pipeline |
| [`rlentity`](rlentity.html) | Stateless entity helpers (move, face, eat, swap, death detection) |
| [`rlgeneration`](rlgeneration.html) | Procedural level and terrain generation |
| [`rlsystems`](rlsystems.html) | Turn-based ECS systems (AI, Initiative, StatusCondition, Door, Cleanup) |
| [`rlasciiclient`](rlasciiclient.html) | Ebiten ASCII client: decodes server snapshots into an entity store and renders a tile-grid viewport |
| [`rltermclient`](rltermclient.html) | Terminal ASCII client (tcell): renders the same snapshots in a headless terminal with GUI overlay support |
| [`rltermgui`](rltermgui.html) | Lightweight terminal GUI layer (tcell): Views, Pane, GUI manager, and drawing helpers |

## Installation

```bash
go get github.com/mechanical-lich/ml-rogue-lib
```

## Dependencies

ML Rogue Lib depends on [MLGE](https://mechanical-lich.github.io/mlge) for the ECS foundation (`ecs.Entity`, `ecs.Component`), dice rolling, and message posting. It also uses `go-perlin` for procedural terrain generation. Pathfinding is provided by the bundled `pkg/path` package — no external pathfinding dependency.

## License

See [LICENSE](https://github.com/mechanical-lich/ml-rogue-lib/blob/main/LICENSE) for details.
