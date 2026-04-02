package rlcomponents

import (
	"testing"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/stretchr/testify/assert"
)

func newStatEntity(ac, str, dex int) *ecs.Entity {
	e := &ecs.Entity{}
	e.AddComponent(&StatsComponent{AC: ac, Str: str, Dex: dex})
	return e
}

// =============================================================================
// Decay
// =============================================================================

func TestStatCondition_DecayCountsDown(t *testing.T) {
	c := &StatConditionComponent{Duration: 3}
	assert.False(t, c.Decay())
	assert.False(t, c.Decay())
	assert.True(t, c.Decay())
}

// =============================================================================
// ApplyOnce
// =============================================================================

func TestStatCondition_ApplyOnce_ModifiesStat(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Hardened",
		Duration: 3,
		Mods:     []StatMod{{Stat: "ac", Delta: 2}},
	}

	c.ApplyOnce(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 12, sc.AC)
}

func TestStatCondition_ApplyOnce_IsIdempotent(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Hardened",
		Duration: 3,
		Mods:     []StatMod{{Stat: "ac", Delta: 2}},
	}

	c.ApplyOnce(e)
	c.ApplyOnce(e)
	c.ApplyOnce(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 12, sc.AC, "applying more than once must not stack the bonus")
}

func TestStatCondition_ApplyOnce_MultipleMods(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Surge",
		Duration: 5,
		Mods: []StatMod{
			{Stat: "str", Delta: 3},
			{Stat: "dex", Delta: -1},
		},
	}

	c.ApplyOnce(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 13, sc.Str)
	assert.Equal(t, 9, sc.Dex)
}

func TestStatCondition_ApplyOnce_NegativeDelta(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Weakened",
		Duration: 2,
		Mods:     []StatMod{{Stat: "str", Delta: -4}},
	}

	c.ApplyOnce(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 6, sc.Str)
}

func TestStatCondition_ApplyOnce_NoStatsComponent_DoesNotPanic(t *testing.T) {
	e := &ecs.Entity{}
	c := &StatConditionComponent{
		Name:     "Hardened",
		Duration: 3,
		Mods:     []StatMod{{Stat: "ac", Delta: 2}},
	}

	assert.NotPanics(t, func() { c.ApplyOnce(e) })
}

// =============================================================================
// Revert
// =============================================================================

func TestStatCondition_Revert_RestoresStat(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Hardened",
		Duration: 1,
		Mods:     []StatMod{{Stat: "ac", Delta: 2}},
	}

	c.ApplyOnce(e)
	c.Revert(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 10, sc.AC, "AC must be restored after Revert")
}

func TestStatCondition_Revert_RestoresMultipleMods(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Surge",
		Duration: 1,
		Mods: []StatMod{
			{Stat: "str", Delta: 3},
			{Stat: "dex", Delta: -1},
		},
	}

	c.ApplyOnce(e)
	c.Revert(e)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 10, sc.Str)
	assert.Equal(t, 10, sc.Dex)
}

func TestStatCondition_Revert_BeforeApply_DoesNotPanic(t *testing.T) {
	e := newStatEntity(10, 10, 10)
	c := &StatConditionComponent{
		Name:     "Hardened",
		Duration: 1,
		Mods:     []StatMod{{Stat: "ac", Delta: 2}},
	}

	assert.NotPanics(t, func() { c.Revert(e) })

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 10, sc.AC, "unapplied revert must not change stats")
}

// =============================================================================
// All supported stat fields
// =============================================================================

func TestStatCondition_AllSupportedStats(t *testing.T) {
	stats := []struct {
		stat string
		get  func(*StatsComponent) int
	}{
		{"ac", func(s *StatsComponent) int { return s.AC }},
		{"str", func(s *StatsComponent) int { return s.Str }},
		{"dex", func(s *StatsComponent) int { return s.Dex }},
		{"con", func(s *StatsComponent) int { return s.Con }},
		{"int", func(s *StatsComponent) int { return s.Int }},
		{"wis", func(s *StatsComponent) int { return s.Wis }},
		{"melee_attack_bonus", func(s *StatsComponent) int { return s.MeleeAttackBonus }},
		{"ranged_attack_bonus", func(s *StatsComponent) int { return s.RangedAttackBonus }},
	}

	for _, tt := range stats {
		t.Run(tt.stat, func(t *testing.T) {
			e := &ecs.Entity{}
			e.AddComponent(&StatsComponent{})
			c := &StatConditionComponent{
				Duration: 3,
				Mods:     []StatMod{{Stat: tt.stat, Delta: 5}},
			}

			c.ApplyOnce(e)
			sc := e.GetComponent(Stats).(*StatsComponent)
			assert.Equal(t, 5, tt.get(sc), "apply: %s", tt.stat)

			c.Revert(e)
			assert.Equal(t, 0, tt.get(sc), "revert: %s", tt.stat)
		})
	}
}

// =============================================================================
// Integration with StatusConditionSystem via ConditionModifier interface
// =============================================================================

func TestStatCondition_ImplementsConditionModifier(t *testing.T) {
	var _ ConditionModifier = &StatConditionComponent{}
}

func TestStatCondition_ImplementsDecayingComponent(t *testing.T) {
	var _ DecayingComponent = &StatConditionComponent{}
}
