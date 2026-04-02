package rlsystems

import (
	"testing"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/stretchr/testify/assert"
)

// --- helpers ---

func newEntity() *ecs.Entity {
	e := &ecs.Entity{}
	e.AddComponent(&rlcomponents.PositionComponent{})
	e.AddComponent(&rlcomponents.MyTurnComponent{})
	return e
}

func withHealth(e *ecs.Entity, hp int) *ecs.Entity {
	e.AddComponent(&rlcomponents.HealthComponent{Health: hp, MaxHealth: hp})
	return e
}

func withBody(e *ecs.Entity, parts ...rlcomponents.BodyPart) *ecs.Entity {
	bc := &rlcomponents.BodyComponent{}
	for _, p := range parts {
		bc.AddPart(p)
	}
	e.AddComponent(bc)
	return e
}

func withSpeed(e *ecs.Entity, speed int) *ecs.Entity {
	e.AddComponent(&rlcomponents.EnergyComponent{Speed: speed})
	return e
}

func tick(sys *StatusConditionSystem, e *ecs.Entity) {
	_ = sys.UpdateEntity(nil, e)
}

// baseSystem returns a StatusConditionSystem with no extras.
func baseSystem() *StatusConditionSystem {
	return &StatusConditionSystem{}
}

// speedModSystem returns a system with Haste and Slowed registered as extras.
func speedModSystem() *StatusConditionSystem {
	return &StatusConditionSystem{
		ExtraStatuses: map[string]ecs.ComponentType{
			"Haste":  rlcomponents.Haste,
			"Slowed": rlcomponents.Slowed,
		},
	}
}

// =============================================================================
// Poisoned
// =============================================================================

func TestPoison_DealsDamageToHealth(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.PoisonedComponent{Duration: 3})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 19, hc.Health)
}

func TestPoison_DealsDamageToBodyPart(t *testing.T) {
	e := withBody(newEntity(), rlcomponents.BodyPart{Name: "torso", HP: 20, MaxHP: 20})
	e.AddComponent(&rlcomponents.PoisonedComponent{Duration: 3})

	tick(baseSystem(), e)

	bc := e.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.Equal(t, 19, bc.Parts["torso"].HP)
}

func TestPoison_ExpiresAfterDuration(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.PoisonedComponent{Duration: 2})
	sys := baseSystem()

	tick(sys, e)
	tick(sys, e)

	assert.False(t, e.HasComponent(rlcomponents.Poisoned), "Poisoned should be removed after duration")
}

func TestPoison_StillActiveBeforeDurationExpires(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.PoisonedComponent{Duration: 3})
	sys := baseSystem()

	tick(sys, e)
	tick(sys, e)

	assert.True(t, e.HasComponent(rlcomponents.Poisoned))
}

// =============================================================================
// Burning
// =============================================================================

func TestBurning_DealsTwoDamagePerTurn(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.BurningComponent{Duration: 3})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 18, hc.Health)
}

func TestBurning_ExpiresAfterDuration(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.BurningComponent{Duration: 1})
	sys := baseSystem()

	tick(sys, e)

	assert.False(t, e.HasComponent(rlcomponents.Burning))
}

// =============================================================================
// Haste — ConditionModifier apply/revert
// =============================================================================

func TestHaste_DoublesSpeed(t *testing.T) {
	e := withSpeed(newEntity(), 100)
	e.AddComponent(&rlcomponents.HasteComponent{Duration: 3})

	tick(speedModSystem(), e)

	ec := e.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	assert.Equal(t, 200, ec.Speed)
}

func TestHaste_DoesNotDoubleSpeedTwice(t *testing.T) {
	e := withSpeed(newEntity(), 100)
	e.AddComponent(&rlcomponents.HasteComponent{Duration: 3})
	sys := speedModSystem()

	tick(sys, e)
	tick(sys, e)

	ec := e.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	assert.Equal(t, 200, ec.Speed)
}

func TestHaste_RevertsSpeedOnExpiry(t *testing.T) {
	e := withSpeed(newEntity(), 100)
	e.AddComponent(&rlcomponents.HasteComponent{Duration: 1})
	sys := speedModSystem()

	tick(sys, e)

	assert.False(t, e.HasComponent(rlcomponents.Haste))
	ec := e.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	assert.Equal(t, 100, ec.Speed, "speed must be restored after Haste expires")
}

// =============================================================================
// Slowed — ConditionModifier apply/revert
// =============================================================================

func TestSlowed_HalvesSpeed(t *testing.T) {
	e := withSpeed(newEntity(), 100)
	e.AddComponent(&rlcomponents.SlowedComponent{Duration: 3})

	tick(speedModSystem(), e)

	ec := e.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	assert.Equal(t, 50, ec.Speed)
}

func TestSlowed_RevertsSpeedOnExpiry(t *testing.T) {
	e := withSpeed(newEntity(), 100)
	e.AddComponent(&rlcomponents.SlowedComponent{Duration: 1})
	sys := speedModSystem()

	tick(sys, e)

	assert.False(t, e.HasComponent(rlcomponents.Slowed))
	ec := e.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	assert.Equal(t, 100, ec.Speed, "speed must be restored after Slowed expires")
}

// =============================================================================
// DamageCondition
// =============================================================================

func TestDamageCondition_DealsDamageToHealth(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.DamageConditionComponent{
		Name:       "Venom",
		Duration:   3,
		DamageDice: "3", // constant 3 damage
		DamageType: "poison",
	})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 17, hc.Health)
}

func TestDamageCondition_DealsDamageToBodyPart(t *testing.T) {
	e := withBody(newEntity(), rlcomponents.BodyPart{Name: "torso", HP: 20, MaxHP: 20})
	e.AddComponent(&rlcomponents.DamageConditionComponent{
		Name:       "Acid Burn",
		Duration:   3,
		DamageDice: "2",
		DamageType: "acid",
	})

	tick(baseSystem(), e)

	bc := e.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.Equal(t, 18, bc.Parts["torso"].HP)
}

func TestDamageCondition_ExpiresAfterDuration(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.DamageConditionComponent{
		Name:       "Venom",
		Duration:   2,
		DamageDice: "1",
	})
	sys := baseSystem()

	tick(sys, e)
	tick(sys, e)

	assert.False(t, e.HasComponent(rlcomponents.DamageCondition))
}

func TestDamageCondition_DiceRollDamageIsWithinRange(t *testing.T) {
	e := withHealth(newEntity(), 1000)
	e.AddComponent(&rlcomponents.DamageConditionComponent{
		Name:       "Venom",
		Duration:   1,
		DamageDice: "1d6",
	})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	dmg := 1000 - hc.Health
	assert.GreaterOrEqual(t, dmg, 1)
	assert.LessOrEqual(t, dmg, 6)
}

// =============================================================================
// ExtraStatuses
// =============================================================================

// chillComponent is a minimal extra status for testing.
type chillComponent struct{ Duration int }

func (c *chillComponent) GetType() ecs.ComponentType { return "Chilled" }
func (c *chillComponent) Decay() bool {
	c.Duration--
	return c.Duration <= 0
}

func TestExtraStatus_CallsOnStatusEffect(t *testing.T) {
	e := newEntity()
	e.AddComponent(&chillComponent{Duration: 3})

	var called []string
	sys := &StatusConditionSystem{
		ExtraStatuses: map[string]ecs.ComponentType{"Chilled": "Chilled"},
		OnStatusEffect: func(entity *ecs.Entity, effectName string) {
			called = append(called, effectName)
		},
	}

	tick(sys, e)

	assert.Contains(t, called, "Chilled")
}

func TestExtraStatus_ExpiresAfterDuration(t *testing.T) {
	e := newEntity()
	e.AddComponent(&chillComponent{Duration: 2})
	sys := &StatusConditionSystem{
		ExtraStatuses: map[string]ecs.ComponentType{"Chilled": "Chilled"},
	}

	tick(sys, e)
	tick(sys, e)

	assert.False(t, e.HasComponent("Chilled"), "extra status should be removed after duration")
}

func TestExtraStatus_OnStatusEffectCalledForBuiltInsAndExtras(t *testing.T) {
	e := withHealth(newEntity(), 20)
	e.AddComponent(&rlcomponents.PoisonedComponent{Duration: 3})
	e.AddComponent(&chillComponent{Duration: 3})

	var called []string
	sys := &StatusConditionSystem{
		ExtraStatuses: map[string]ecs.ComponentType{"Chilled": "Chilled"},
		OnStatusEffect: func(_ *ecs.Entity, effectName string) {
			called = append(called, effectName)
		},
	}

	tick(sys, e)

	assert.Contains(t, called, "Poisoned")
	assert.Contains(t, called, "Chilled")
}

// =============================================================================
// Regeneration
// =============================================================================

func TestRegeneration_HealsHealth(t *testing.T) {
	e := newEntity()
	e.AddComponent(&rlcomponents.HealthComponent{Health: 10, MaxHealth: 20})
	e.AddComponent(&rlcomponents.RegenerationComponent{Amount: 3})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 13, hc.Health)
}

func TestRegeneration_DoesNotExceedMaxHealth(t *testing.T) {
	e := newEntity()
	e.AddComponent(&rlcomponents.HealthComponent{Health: 19, MaxHealth: 20})
	e.AddComponent(&rlcomponents.RegenerationComponent{Amount: 5})

	tick(baseSystem(), e)

	hc := e.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
	assert.Equal(t, 20, hc.Health)
}

func TestRegeneration_HealsBodyPart(t *testing.T) {
	e := withBody(newEntity(), rlcomponents.BodyPart{Name: "arm", HP: 5, MaxHP: 10})
	e.AddComponent(&rlcomponents.RegenerationComponent{Amount: 2})

	tick(baseSystem(), e)

	bc := e.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.Equal(t, 7, bc.Parts["arm"].HP)
}

func TestRegeneration_DoesNotHealAmputatedBodyPart(t *testing.T) {
	e := withBody(newEntity(), rlcomponents.BodyPart{Name: "arm", HP: 0, MaxHP: 10, Amputated: true})
	e.AddComponent(&rlcomponents.RegenerationComponent{Amount: 5})

	tick(baseSystem(), e)

	bc := e.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.Equal(t, 0, bc.Parts["arm"].HP, "amputated parts should not regenerate")
}

func TestRegeneration_UnbreaksBrokenBodyPartWhenHPAboveZero(t *testing.T) {
	e := withBody(newEntity(), rlcomponents.BodyPart{Name: "arm", HP: 0, MaxHP: 10, Broken: true})
	e.AddComponent(&rlcomponents.RegenerationComponent{Amount: 3})

	tick(baseSystem(), e)

	bc := e.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
	assert.False(t, bc.Parts["arm"].Broken, "part should no longer be broken once HP > 0")
}
