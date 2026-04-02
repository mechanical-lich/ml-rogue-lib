package rlcomponents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDamageCondition_DecayCountsDown(t *testing.T) {
	c := &DamageConditionComponent{Duration: 3}
	assert.False(t, c.Decay())
	assert.False(t, c.Decay())
	assert.True(t, c.Decay())
}

func TestDamageCondition_Roll_Constant(t *testing.T) {
	c := &DamageConditionComponent{DamageDice: "5"}
	assert.Equal(t, 5, c.Roll())
}

func TestDamageCondition_Roll_DiceInRange(t *testing.T) {
	c := &DamageConditionComponent{DamageDice: "1d6"}
	for range 50 {
		r := c.Roll()
		assert.GreaterOrEqual(t, r, 1)
		assert.LessOrEqual(t, r, 6)
	}
}

func TestDamageCondition_Roll_InvalidDice_ReturnsOne(t *testing.T) {
	c := &DamageConditionComponent{DamageDice: "not-a-dice"}
	assert.Equal(t, 1, c.Roll())
}

func TestDamageCondition_Roll_MinimumOne(t *testing.T) {
	// A dice expression that could theoretically roll 0 should clamp to 1.
	// "1d1-1" rolls 0; Roll() must return 1.
	c := &DamageConditionComponent{DamageDice: "1d1-1"}
	assert.Equal(t, 1, c.Roll())
}

func TestDamageCondition_ImplementsDecayingComponent(t *testing.T) {
	var _ DecayingComponent = &DamageConditionComponent{}
}
