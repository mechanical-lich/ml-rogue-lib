package rlsystems

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlai"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlentity"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/path"
	"github.com/mechanical-lich/mlge/utility"
)

// AISystem handles WanderAI, HostileAI, and DefensiveAI each turn.
//
// Extension hooks:
//   - HostileTargetMatch: determines whether a candidate entity is a valid target
//     for a hostile entity. Return true to target it. If nil, defaults to targeting
//     any entity with Health that is not on the same faction.
//   - GetPath: pathfinding function used by HostileAI to plan movement.
//     Signature mirrors mlge/path.AStar usage. If nil, hostile AI wanders when
//     no simple delta movement is available.
//   - OnWander: called when an entity's WanderAI triggers. Use this to add custom
//     wander logic (e.g., play footstep sounds, update bookkeeping).
//   - OnHostileAttack: called when a hostile entity attacks. Use this to add
//     custom hit logic beyond the built-in combat.Hit call.
type AISystem struct {
	// HostileTargetMatch decides if candidate is a valid target for self.
	// Defaults to: has Health, different faction, not Dead.
	HostileTargetMatch func(self, candidate *ecs.Entity) bool

	// GetPath returns a path from fromTile to toTile, reusing the provided slice.
	// If nil, HostileAI falls back to direct delta movement (no pathfinding).
	GetPath func(level rlworld.LevelInterface, from, to rlworld.TileInterface, reuse []path.Pather) []path.Pather

	// OnWander is called after a WanderAI move.
	OnWander func(entity *ecs.Entity)

	// OnHostileAttack is called when a hostile entity hits a target.
	OnHostileAttack func(level rlworld.LevelInterface, attacker, target *ecs.Entity)

	entitiesHitBuf []*ecs.Entity
	hostileSelf    *ecs.Entity
}

func NewAISystem() *AISystem {
	return &AISystem{
		entitiesHitBuf: make([]*ecs.Entity, 0, 5),
	}
}

var aiRequires = []ecs.ComponentType{rlcomponents.Position, rlcomponents.MyTurn}

func (s *AISystem) Requires() []ecs.ComponentType {
	return aiRequires
}

func (s *AISystem) UpdateSystem(data interface{}) error {
	return nil
}

func (s *AISystem) UpdateEntity(levelInterface interface{}, entity *ecs.Entity) error {
	level := levelInterface.(rlworld.LevelInterface)

	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}

	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

	if rlentity.HandleDeath(entity) {
		return nil
	}

	// ── Wander AI ─────────────────────────────────────────────────────────────
	if entity.HasComponent(rlcomponents.WanderAI) {
		dx := utility.GetRandom(-1, 2)
		dy := 0
		if dx == 0 {
			dy = utility.GetRandom(-1, 2)
		}
		rlentity.Move(entity, level, dx, dy, 0)
		rlentity.Face(entity, dx, dy)
		if s.OnWander != nil {
			s.OnWander(entity)
		}
	}

	// ── Hostile AI ────────────────────────────────────────────────────────────
	if entity.HasComponent(rlcomponents.HostileAI) {
		hc := entity.GetComponent(rlcomponents.HostileAI).(*rlcomponents.HostileAIComponent)
		dx, dy := 0, 0

		matchFn := s.hostileMatchFor(entity)
		closest := level.GetClosestEntityMatching(
			pc.GetX(), pc.GetY(), pc.GetZ(),
			hc.SightRange, hc.SightRange,
			entity, matchFn,
		)

		if closest != nil {
			targetPC := closest.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			hc.TargetX = targetPC.GetX()
			hc.TargetY = targetPC.GetY()

			if s.GetPath != nil {
				from := level.GetTileAt(pc.GetX(), pc.GetY(), pc.GetZ())
				to := level.GetTileAt(hc.TargetX, hc.TargetY, pc.GetZ())
				if from != nil && to != nil {
					needNew := len(hc.Path) == 0
					if !needNew {
						last := hc.Path[len(hc.Path)-1].(rlworld.TileInterface)
						lx, ly, _ := last.Coords()
						needNew = lx != hc.TargetX || ly != hc.TargetY
					}
					if needNew {
						hc.Path = s.GetPath(level, from, to, hc.Path)
					}
				}
			}

			if len(hc.Path) > 0 {
				t := hc.Path[0].(rlworld.TileInterface)
				tx, ty, _ := t.Coords()
				if pc.GetX() == tx && pc.GetY() == ty {
					hc.Path = hc.Path[1:]
				}
				if len(hc.Path) > 0 {
					t = hc.Path[0].(rlworld.TileInterface)
					tx, ty, _ = t.Coords()
					if pc.GetX() < tx {
						dx = 1
					} else if pc.GetX() > tx {
						dx = -1
					}
					if pc.GetY() < ty {
						dy = 1
					} else if pc.GetY() > ty {
						dy = -1
					}
				}
			} else {
				// No path — move directly toward target.
				dx, dy, _ = rlai.TrackTarget(pc.GetX(), pc.GetY(), pc.GetZ(), hc.TargetX, hc.TargetY, pc.GetZ())
			}
		}

		// Random wander when no target.
		if dx == 0 && dy == 0 {
			dx = utility.GetRandom(-1, 2)
			if dx == 0 {
				dy = utility.GetRandom(-1, 2)
			}
		}

		hit := false
		s.entitiesHitBuf = s.entitiesHitBuf[:0]
		level.GetEntitiesAt(pc.GetX()+dx, pc.GetY()+dy, pc.GetZ(), &s.entitiesHitBuf)
		for _, e := range s.entitiesHitBuf {
			if e != entity && e.HasComponent(rlcomponents.Health) {
				rlcombat.Hit(level, entity, e, true)
				rlentity.Eat(entity, e)
				if s.OnHostileAttack != nil {
					s.OnHostileAttack(level, entity, e)
				}
				hit = true
			}
		}
		if !hit {
			rlentity.Move(entity, level, dx, dy, 0)
		}
		rlentity.Face(entity, dx, dy)
	}

	// ── Defensive AI ──────────────────────────────────────────────────────────
	if entity.HasComponent(rlcomponents.DefensiveAI) {
		aic := entity.GetComponent(rlcomponents.DefensiveAI).(*rlcomponents.DefensiveAIComponent)
		if aic.Attacked {
			attacker := level.GetSolidEntityAt(aic.AttackerX, aic.AttackerY, pc.GetZ())
			if attacker == nil {
				aic.Attacked = false
			} else {
				rlcombat.Hit(level, entity, attacker, true)
			}
			dx := 0
			dy := 0
			if pc.GetX() < aic.AttackerX {
				dx = 1
			} else if pc.GetX() > aic.AttackerX {
				dx = -1
			}
			if pc.GetY() < aic.AttackerY {
				dy = 1
			} else if pc.GetY() > aic.AttackerY {
				dy = -1
			}
			rlentity.Face(entity, dx, dy)
		}
	}

	return nil
}

// hostileMatchFor returns the target-match function, using the custom hook or
// the built-in default (targets any entity with Health on a different faction).
func (s *AISystem) hostileMatchFor(self *ecs.Entity) func(*ecs.Entity) bool {
	if s.HostileTargetMatch != nil {
		return func(candidate *ecs.Entity) bool {
			return s.HostileTargetMatch(self, candidate)
		}
	}
	return func(candidate *ecs.Entity) bool {
		if candidate.HasComponent(rlcomponents.Dead) {
			return false
		}
		if !candidate.HasComponent(rlcomponents.Health) {
			return false
		}
		if rlcombat.IsFriendly(self, candidate) {
			return false
		}
		return true
	}
}
