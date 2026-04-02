package rlsystems

import (
	"math/rand"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// applyStatusDamage deals dmg to a random non-amputated body part, or directly
// to Health if the entity has no BodyComponent.
func applyStatusDamage(entity *ecs.Entity, dmg int) {
	if entity.HasComponent(rlcomponents.Body) {
		bc := entity.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
		var available []string
		for name, part := range bc.Parts {
			if !part.Amputated {
				available = append(available, name)
			}
		}
		if len(available) > 0 {
			name := available[rand.Intn(len(available))]
			part := bc.Parts[name]
			part.HP -= dmg
			if part.HP <= 0 && !part.Broken {
				part.Broken = true
			}
			bc.Parts[name] = part
		} else {
			entity.AddComponent(&rlcomponents.DeadComponent{})
		}
	} else if entity.HasComponent(rlcomponents.Health) {
		hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
		hc.Health -= dmg
	}
}

// statusEntry pairs a decaying component type with its per-turn effect name.
type statusEntry struct {
	componentType ecs.ComponentType
	effectName    string
}

// StatusConditionSystem ticks decaying status effects (Poisoned, Burning, Alerted)
// and applies their per-turn damage. It also handles Regeneration and LightSensitive.
//
// Extension hooks:
//   - ExtraStatuses: additional DecayingComponent types to tick alongside built-ins.
//     Map key is the effect name passed to OnStatusEffect.
//   - OnStatusEffect: called each turn for every active status effect on an entity.
//     Use this to apply custom damage, spawn FX entities, play sounds, etc.
//     The built-in effects (Poisoned: -1 HP, Burning: -2 HP) run before this hook.
type StatusConditionSystem struct {
	// ExtraStatuses lets games register additional decaying component types.
	// Key = effect name, Value = component type.
	ExtraStatuses map[string]ecs.ComponentType

	// OnStatusEffect is called for each active status each turn.
	// effectName is one of "Poisoned", "Burning", "Alerted", or a key from ExtraStatuses.
	OnStatusEffect func(entity *ecs.Entity, effectName string)
}

var statusConditionRequires = []ecs.ComponentType{
	rlcomponents.Position,
	rlcomponents.MyTurn,
}

func (s *StatusConditionSystem) Requires() []ecs.ComponentType {
	return statusConditionRequires
}

func (s *StatusConditionSystem) UpdateSystem(data any) error {
	return nil
}

func (s *StatusConditionSystem) UpdateEntity(levelInterface any, entity *ecs.Entity) error {
	statuses := []statusEntry{
		{rlcomponents.Poisoned, "Poisoned"},
		{rlcomponents.Alerted, "Alerted"},
		{rlcomponents.Burning, "Burning"},
		{rlcomponents.StatCondition, "StatCondition"},
		{rlcomponents.DamageCondition, "DamageCondition"},
	}
	for name, ct := range s.ExtraStatuses {
		statuses = append(statuses, statusEntry{ct, name})
	}

	for _, se := range statuses {
		if !entity.HasComponent(se.componentType) {
			continue
		}
		dc := entity.GetComponent(se.componentType).(rlcomponents.DecayingComponent)

		// Apply speed-modifying effects once.
		if sm, ok := dc.(rlcomponents.ConditionModifier); ok {
			sm.ApplyOnce(entity)
		}

		if dc.Decay() {
			// Revert speed-modifying effects before removing.
			if sm, ok := dc.(rlcomponents.ConditionModifier); ok {
				sm.Revert(entity)
			}
			entity.RemoveComponent(se.componentType)
		}

		// Built-in damage effects.
		var dmg int
		switch se.effectName {
		case "Poisoned":
			dmg = 1
		case "Burning":
			dmg = 2
		case "DamageCondition":
			dmg = dc.(*rlcomponents.DamageConditionComponent).Roll()
		}
		if dmg > 0 {
			applyStatusDamage(entity, dmg)
		}

		if s.OnStatusEffect != nil {
			s.OnStatusEffect(entity, se.effectName)
		}
	}

	// Regeneration.
	if entity.HasComponent(rlcomponents.Regeneration) {
		rc := entity.GetComponent(rlcomponents.Regeneration).(*rlcomponents.RegenerationComponent)
		if entity.HasComponent(rlcomponents.Body) {
			bc := entity.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
			for name, part := range bc.Parts {
				if !part.Amputated && part.HP < part.MaxHP {
					part.HP += rc.Amount
					if part.HP > part.MaxHP {
						part.HP = part.MaxHP
					}
					if part.HP > 0 && part.Broken {
						part.Broken = false
					}
					bc.Parts[name] = part
				}
			}
		} else if entity.HasComponent(rlcomponents.Health) {
			hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
			hc.Health += rc.Amount
			if hc.Health > hc.MaxHealth {
				hc.Health = hc.MaxHealth
			}
		}
	}

	return nil
}
