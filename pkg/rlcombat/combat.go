package rlcombat

import (
	"fmt"
	"log"
	"strings"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/mlge/utility"
)

const DefaultDamageType = "bludgeoning"

// GetModifier returns the D&D-style ability modifier for a stat value.
func GetModifier(stat int) int {
	return (stat - 10) / 2
}

// IsFriendly returns true if attacker and defender share the same non-empty faction.
func IsFriendly(attacker, defender *ecs.Entity) bool {
	if !attacker.HasComponent(rlcomponents.Description) || !defender.HasComponent(rlcomponents.Description) {
		return false
	}
	a := attacker.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
	d := defender.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent)
	return a.Faction != "" && a.Faction != "none" && a.Faction == d.Faction
}

// TriggerDefenses alerts the defender's AI components that they were attacked.
func TriggerDefenses(defender *ecs.Entity, attackerX, attackerY int) {
	if defender.HasComponent(rlcomponents.DefensiveAI) {
		daic := defender.GetComponent(rlcomponents.DefensiveAI).(*rlcomponents.DefensiveAIComponent)
		daic.Attacked = true
		daic.AttackerX = attackerX
		daic.AttackerY = attackerY
	}
	if defender.HasComponent(rlcomponents.AIMemory) {
		mem := defender.GetComponent(rlcomponents.AIMemory).(*rlcomponents.AIMemoryComponent)
		mem.Attacked = true
		mem.AttackerX = attackerX
		mem.AttackerY = attackerY
	}
	if !defender.HasComponent(rlcomponents.Alerted) {
		defender.AddComponent(&rlcomponents.AlertedComponent{Duration: 120})
	}
}

// ApplyStatusEffects transfers status effects from attacker to defender on a successful hit.
func ApplyStatusEffects(attacker, defender *ecs.Entity) {
	if attacker.HasComponent(rlcomponents.Poisonous) && !defender.HasComponent(rlcomponents.Poisoned) {
		pc := attacker.GetComponent(rlcomponents.Poisonous).(*rlcomponents.PoisonousComponent)
		defender.AddComponent(&rlcomponents.PoisonedComponent{Duration: pc.Duration})
	}
}

// GetAttackDice returns (dice string, damage type, modifier) for an entity's attack.
func GetAttackDice(entity *ecs.Entity) (string, string, int) {
	sc := entity.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	attackDice := sc.BasicAttackDice
	damageType := sc.BaseDamageType
	if damageType == "" {
		damageType = DefaultDamageType
	}
	mod := GetModifier(sc.Str)

	if entity.HasComponent(rlcomponents.Inventory) {
		inv := entity.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent)
		mod += inv.GetAttackModifier()
		if d := inv.GetAttackDice(); d != "" {
			attackDice = d
		}
		if inv.RightHand != nil && inv.RightHand.HasComponent(rlcomponents.Weapon) {
			w := inv.RightHand.GetComponent(rlcomponents.Weapon).(*rlcomponents.WeaponComponent)
			if w.DamageType != "" {
				damageType = w.DamageType
			}
		}
	}
	return attackDice, damageType, mod
}

// IsInArrowPath returns true if (tX,tY) is reachable from (aX,aY) in a straight
// or diagonal line within maxRange tiles.
func IsInArrowPath(aX, aY, tX, tY, maxRange int) bool {
	dx := tX - aX
	dy := tY - aY
	if dx == 0 && dy != 0 && utility.Abs(dy) <= maxRange {
		return true
	}
	if dy == 0 && dx != 0 && utility.Abs(dx) <= maxRange {
		return true
	}
	if utility.Abs(dx) == utility.Abs(dy) && utility.Abs(dx) <= maxRange {
		return true
	}
	return false
}

// InflictDamage rolls and applies damage from attacker to defender,
// accounting for resistances and weaknesses.
func InflictDamage(attacker, defender *ecs.Entity) {
	attackDice, damageType, mod := GetAttackDice(attacker)

	d := attackDice
	if strings.Contains(d, "d") {
		if mod >= 0 {
			d = fmt.Sprintf("%s+%d", attackDice, mod)
		} else {
			d = fmt.Sprintf("%s%d", attackDice, mod)
		}
	}

	damage := 0
	roll, err := dice.ParseDiceRequest(d)
	if err == nil {
		damage = roll.Result
	} else {
		log.Print("rlcombat: error rolling dice: ", d)
	}

	// Resistances and weaknesses.
	if hasResistance(defender, damageType) {
		damage /= 2
	}
	if hasWeakness(defender, damageType) {
		damage *= 2
	}
	if damage <= 0 {
		damage = 1
	}

	if defender.HasComponent(rlcomponents.Health) {
		hc := defender.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
		if hc.Health > 0 {
			hc.Health -= damage
			if attacker.HasComponent(rlcomponents.Description) && defender.HasComponent(rlcomponents.Description) {
				atkName := attacker.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
				defName := defender.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
				message.PostTaggedMessage("combat", atkName, fmt.Sprintf("hit %s for %d (%s)", defName, damage, damageType))
			}
		}
	}
}

// Hit performs a full D&D-style melee attack from entity to entityHit.
// If swap is true and the two entities are friendly, they swap positions instead.
func Hit(level rlworld.LevelInterface, entity, entityHit *ecs.Entity, swap bool) {
	if entity == nil || entityHit == nil || entity == entityHit {
		return
	}

	if IsFriendly(entity, entityHit) {
		if swap {
			pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			hitPC := entityHit.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			oldX, oldY, oldZ := pc.GetX(), pc.GetY(), pc.GetZ()
			level.PlaceEntity(hitPC.GetX(), hitPC.GetY(), hitPC.GetZ(), entity)
			level.PlaceEntity(oldX, oldY, oldZ, entityHit)
		}
		return
	}

	if !entityHit.HasComponent(rlcomponents.Health) || !entityHit.HasComponent(rlcomponents.Stats) || !entity.HasComponent(rlcomponents.Stats) {
		return
	}

	sc := entity.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	hitSc := entityHit.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)

	mod := GetModifier(sc.Dex)
	if entity.HasComponent(rlcomponents.Inventory) {
		mod += entity.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent).GetAttackModifier()
	}

	acBonus := 0
	if entityHit.HasComponent(rlcomponents.Inventory) {
		acBonus += entityHit.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent).GetDefenseModifier()
	}

	roll, err := dice.ParseDiceRequest("1d20")
	if err != nil {
		log.Print("rlcombat: error rolling d20: ", err)
		return
	}

	if roll.Result+mod > hitSc.AC+acBonus {
		InflictDamage(entity, entityHit)
		ApplyStatusEffects(entity, entityHit)
	} else {
		if entity.HasComponent(rlcomponents.Description) && entityHit.HasComponent(rlcomponents.Description) {
			atkName := entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
			defName := entityHit.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
			message.PostTaggedMessage("combat", atkName, fmt.Sprintf("missed %s", defName))
		}
	}

	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	TriggerDefenses(entityHit, pc.GetX(), pc.GetY())
}

func hasResistance(defender *ecs.Entity, damageType string) bool {
	if !defender.HasComponent(rlcomponents.Stats) {
		return false
	}
	sc := defender.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	for _, r := range sc.Resistances {
		if r == damageType {
			return true
		}
	}
	if defender.HasComponent(rlcomponents.Inventory) {
		inv := defender.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent)
		for _, slot := range []*ecs.Entity{inv.Head, inv.Torso, inv.Legs, inv.Feet, inv.LeftHand, inv.RightHand} {
			if slot != nil && slot.HasComponent(rlcomponents.Armor) {
				armor := slot.GetComponent(rlcomponents.Armor).(*rlcomponents.ArmorComponent)
				for _, r := range armor.Resistances {
					if r == damageType {
						return true
					}
				}
			}
		}
	}
	return false
}

func hasWeakness(defender *ecs.Entity, damageType string) bool {
	if !defender.HasComponent(rlcomponents.Stats) {
		return false
	}
	sc := defender.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	for _, w := range sc.Weaknesses {
		if w == damageType {
			return true
		}
	}
	return false
}
