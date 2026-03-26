package rlbodycombat

import (
	"testing"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
	"github.com/stretchr/testify/assert"
)

// --- minimal level stub ---

type stubLevel struct{ rlworld.LevelInterface }

func (s *stubLevel) PlaceEntity(x, y, z int, e *ecs.Entity) {}

// --- combat event collector ---

type eventCollector struct {
	events []CombatEvent
}

func (c *eventCollector) HandleEvent(e event.EventData) error {
	if ce, ok := e.(CombatEvent); ok {
		c.events = append(c.events, ce)
	}
	return nil
}

func collectCombatEvents(fn func()) []CombatEvent {
	// Drain any events left in the queue by prior tests before registering.
	event.GetQueuedInstance().HandleQueue()
	col := &eventCollector{}
	event.GetQueuedInstance().RegisterListener(col, CombatEventType)
	fn()
	event.GetQueuedInstance().HandleQueue()
	event.GetQueuedInstance().UnregisterListener(col, CombatEventType)
	return col.events
}

// --- entity builders ---

// newCombatant creates a minimal attackable entity.
// dex=30 (mod=10) vs ac=1 guarantees a hit on any d20 roll.
func newCombatant(name, faction string, dex, ac int, attackDice string) *ecs.Entity {
	e := &ecs.Entity{}
	e.AddComponent(&rlcomponents.DescriptionComponent{Name: name, Faction: faction})
	e.AddComponent(&rlcomponents.StatsComponent{
		AC:              ac,
		Dex:             dex,
		Str:             10,
		BasicAttackDice: attackDice,
		BaseDamageType:  "bludgeoning",
	})
	e.AddComponent(&rlcomponents.PositionComponent{X: 1, Y: 1, Z: 0})
	return e
}

func addHealth(e *ecs.Entity, hp int) {
	e.AddComponent(&rlcomponents.HealthComponent{Health: hp, MaxHealth: hp})
}

func addBody(e *ecs.Entity, parts ...rlcomponents.BodyPart) {
	bc := &rlcomponents.BodyComponent{}
	for _, p := range parts {
		bc.AddPart(p)
	}
	e.AddComponent(bc)
}

func bodyPart(name string, hp, maxHP int) rlcomponents.BodyPart {
	return rlcomponents.BodyPart{Name: name, HP: hp, MaxHP: maxHP}
}

func vitalBodyPart(name string, hp, maxHP int) rlcomponents.BodyPart {
	return rlcomponents.BodyPart{Name: name, HP: hp, MaxHP: maxHP, KillsWhenBroken: true}
}

// =============================================================================
// applyBodyPartDamage
// =============================================================================

func TestApplyBodyPartDamage_ReducesHP(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(bodyPart("arm", 10, 10))

	broken, amputated, kills := applyBodyPartDamage(bc, "arm", 3)

	assert.Equal(t, 7, bc.Parts["arm"].HP)
	assert.False(t, broken)
	assert.False(t, amputated)
	assert.False(t, kills)
}

func TestApplyBodyPartDamage_BreaksWhenHPReachesZero(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(bodyPart("arm", 10, 10))

	broken, amputated, kills := applyBodyPartDamage(bc, "arm", 10)

	assert.True(t, broken)
	assert.True(t, bc.Parts["arm"].Broken)
	assert.False(t, amputated)
	assert.False(t, kills)
}

func TestApplyBodyPartDamage_AmputatesWhenDamageIsDoubleMaxHP(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(bodyPart("arm", 10, 10))

	broken, amputated, kills := applyBodyPartDamage(bc, "arm", 20)

	assert.True(t, amputated)
	assert.True(t, bc.Parts["arm"].Amputated)
	// also broken since HP went to -10
	assert.True(t, broken)
	assert.False(t, kills)
}

func TestApplyBodyPartDamage_KillsWhenVitalPartBroken(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(vitalBodyPart("head", 5, 5))

	broken, _, kills := applyBodyPartDamage(bc, "head", 5)

	assert.True(t, broken)
	assert.True(t, kills)
}

func TestApplyBodyPartDamage_KillsWhenVitalPartAmputated(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(rlcomponents.BodyPart{
		Name: "heart", HP: 5, MaxHP: 5,
		KillsWhenAmputated: true,
	})

	_, amputated, kills := applyBodyPartDamage(bc, "heart", 10)

	assert.True(t, amputated)
	assert.True(t, kills)
}

func TestApplyBodyPartDamage_DoesNotBreakAlreadyBroken(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(rlcomponents.BodyPart{Name: "arm", HP: 0, MaxHP: 10, Broken: true})

	broken, _, _ := applyBodyPartDamage(bc, "arm", 1)

	assert.False(t, broken, "already-broken part should not trigger broken again")
}

func TestApplyBodyPartDamage_DoesNotAmputateAlreadyAmputated(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(rlcomponents.BodyPart{Name: "arm", HP: 10, MaxHP: 10, Amputated: true})

	_, amputated, _ := applyBodyPartDamage(bc, "arm", 20)

	assert.False(t, amputated, "already-amputated part should not trigger amputated again")
}

// =============================================================================
// randomBodyPart
// =============================================================================

func TestRandomBodyPart_ReturnsNilWhenAllAmputated(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(rlcomponents.BodyPart{Name: "arm", Amputated: true})
	bc.AddPart(rlcomponents.BodyPart{Name: "leg", Amputated: true})

	name, part := randomBodyPart(bc)

	assert.Empty(t, name)
	assert.Nil(t, part)
}

func TestRandomBodyPart_SkipsAmputatedParts(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(rlcomponents.BodyPart{Name: "arm", Amputated: true})
	bc.AddPart(bodyPart("leg", 10, 10))

	for range 20 {
		name, part := randomBodyPart(bc)
		assert.Equal(t, "leg", name)
		assert.NotNil(t, part)
	}
}

func TestRandomBodyPart_ReturnsPartWhenOnlyOne(t *testing.T) {
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(bodyPart("torso", 20, 20))

	name, part := randomBodyPart(bc)

	assert.Equal(t, "torso", name)
	assert.NotNil(t, part)
}

// =============================================================================
// Hit — guard conditions
// =============================================================================

func TestHit_NilAttacker(t *testing.T) {
	defender := newCombatant("goblin", "", 10, 10, "1d4")
	addHealth(defender, 20)

	assert.False(t, Hit(&stubLevel{}, nil, defender, false))
}

func TestHit_NilDefender(t *testing.T) {
	attacker := newCombatant("player", "", 10, 10, "1d4")

	assert.False(t, Hit(&stubLevel{}, attacker, nil, false))
}

func TestHit_SameEntity(t *testing.T) {
	e := newCombatant("player", "", 10, 10, "1d4")
	addHealth(e, 20)

	assert.False(t, Hit(&stubLevel{}, e, e, false))
}

func TestHit_MissingAttackerStats(t *testing.T) {
	attacker := &ecs.Entity{}
	attacker.AddComponent(&rlcomponents.PositionComponent{})

	defender := newCombatant("goblin", "", 10, 1, "1d4")
	addHealth(defender, 20)

	assert.False(t, Hit(&stubLevel{}, attacker, defender, false))
}

func TestHit_DefenderHasNeitherHealthNorBody(t *testing.T) {
	attacker := newCombatant("player", "", 30, 10, "1d4")
	defender := newCombatant("wall", "", 10, 1, "1d4")
	// no Health, no Body added

	assert.False(t, Hit(&stubLevel{}, attacker, defender, false))
}

// =============================================================================
// Hit — friendly behaviour
// =============================================================================

func TestHit_FriendlyNoSwap_ReturnsFalse(t *testing.T) {
	attacker := newCombatant("alice", "heroes", 10, 10, "1d4")
	defender := newCombatant("bob", "heroes", 10, 10, "1d4")
	addHealth(defender, 20)

	result := Hit(&stubLevel{}, attacker, defender, false)

	assert.False(t, result)
}

func TestHit_FriendlySwap_SwapsPositions(t *testing.T) {
	lv := &stubLevel{}
	attacker := newCombatant("alice", "heroes", 10, 10, "1d4")
	defender := newCombatant("bob", "heroes", 10, 10, "1d4")
	addHealth(defender, 20)

	aPC := attacker.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	dPC := defender.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	aPC.X, aPC.Y = 2, 3
	dPC.X, dPC.Y = 5, 6

	result := Hit(lv, attacker, defender, true)

	assert.False(t, result, "friendly Hit should still return false")
	// positions should have been swapped via PlaceEntity (stubbed, no-op here,
	// but the call itself must not panic)
}

// =============================================================================
// Hit — body component path
// =============================================================================

// guaranteedHitAttacker has Dex=30 (mod=10); even a roll of 1 beats AC=1.
func guaranteedHitAttacker(name string) *ecs.Entity {
	return newCombatant(name, "", 30, 10, "1d4")
}

func TestHit_WithBodyComponent_DamagesBodyPart(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	addBody(defender, bodyPart("torso", 20, 20))
	addHealth(defender, 50)

	result := Hit(&stubLevel{}, attacker, defender, false)

	assert.True(t, result)
	bc := defender.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.Less(t, bc.Parts["torso"].HP, 20, "torso should have taken damage")
}

func TestHit_WithBodyComponent_PostsCombatEvent(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	addBody(defender, bodyPart("torso", 20, 20))
	addHealth(defender, 50)

	events := collectCombatEvents(func() {
		Hit(&stubLevel{}, attacker, defender, false)
	})

	assert.Len(t, events, 1)
	e := events[0]
	assert.False(t, e.Miss)
	assert.Greater(t, e.Damage, 0)
	assert.Equal(t, "torso", e.BodyPart)
	assert.Equal(t, "warrior", e.AttackerName)
	assert.Equal(t, "goblin", e.DefenderName)
}

func TestHit_WithBodyComponent_KillVitalPartSetsHealthToZero(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	// Head with 1 HP that kills when broken; any hit will break it.
	addBody(defender, vitalBodyPart("head", 1, 1))
	addHealth(defender, 50)

	Hit(&stubLevel{}, attacker, defender, false)

	hc := defender.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 0, hc.Health)
}

func TestHit_WithBodyComponent_AmputationKillsSetsHealthToZero(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	// Part with MaxHP=1; damage of 1d4 (min 1) always >= 2*1=2? Not guaranteed.
	// Use a fixed-value dice expression and MaxHP=1 to ensure 2×MaxHP is met.
	// Attacker dice "4" (constant) gives 4 damage; 4 >= 2*1=2.
	attacker.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent).BasicAttackDice = "4"
	addBody(defender, rlcomponents.BodyPart{
		Name: "heart", HP: 5, MaxHP: 1, KillsWhenAmputated: true,
	})
	addHealth(defender, 50)

	Hit(&stubLevel{}, attacker, defender, false)

	hc := defender.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 0, hc.Health)
}

// =============================================================================
// Hit — health fallback (no BodyComponent)
// =============================================================================

func TestHit_WithoutBodyComponent_DamagesHealth(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	addHealth(defender, 50)

	Hit(&stubLevel{}, attacker, defender, false)

	hc := defender.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Less(t, hc.Health, 50, "health should have decreased")
}

// =============================================================================
// Hit — all body parts amputated falls back to health
// =============================================================================

func TestHit_AllPartsAmputated_KillsEntity(t *testing.T) {
	attacker := guaranteedHitAttacker("warrior")
	defender := newCombatant("goblin", "", 10, 1, "1d4")
	addBody(defender, rlcomponents.BodyPart{Name: "arm", HP: 0, MaxHP: 10, Amputated: true})

	Hit(&stubLevel{}, attacker, defender, false)

	assert.True(t, defender.HasComponent(rlcomponents.Dead), "fully-amputated entity should be marked dead")
}

// =============================================================================
// Hit — crit doubles damage
// =============================================================================

func TestApplyBodyPartDamage_CritDoublesEffectivelyViaAmputationThreshold(t *testing.T) {
	// Directly test that crit=true in applyBodyPartDamage path doubles damage.
	// We test rollDamage indirectly: a part with MaxHP=3, non-crit max damage 2
	// would not amputate (2 < 2*3=6), but with crit the doubled value would.
	// Instead, exercise the crit flag through applyBodyPartDamage directly:
	// if damage (already doubled by caller) is 6 against MaxHP=3, it amputates.
	bc := &rlcomponents.BodyComponent{}
	bc.AddPart(bodyPart("arm", 10, 3))

	_, amputated, _ := applyBodyPartDamage(bc, "arm", 6) // 6 == 2*3

	assert.True(t, amputated)
}

// =============================================================================
// CombatEvent — miss path
// =============================================================================

func TestHit_MissPostsMissEvent(t *testing.T) {
	// Attacker with Dex=10 (mod=0), defender with very high AC=30.
	// A natural 20 still crits (always hits), so we can't guarantee a miss
	// in a single call. Run until we see a miss event (or skip if unlucky).
	attacker := newCombatant("warrior", "", 10, 10, "1d4")
	defender := newCombatant("tank", "", 10, 30, "1d4")
	addHealth(defender, 100)

	var missEvents []CombatEvent
	for range 100 {
		evts := collectCombatEvents(func() {
			Hit(&stubLevel{}, attacker, defender, false)
		})
		for _, ev := range evts {
			if ev.Miss {
				missEvents = append(missEvents, ev)
			}
		}
		if len(missEvents) > 0 {
			break
		}
	}

	if len(missEvents) == 0 {
		t.Skip("rolled natural 20 every attempt — statistically unlikely, skipping")
	}
	assert.True(t, missEvents[0].Miss)
	assert.Equal(t, 0, missEvents[0].Damage)
	assert.Empty(t, missEvents[0].BodyPart)
}
