package rlcomponents

import (
	"fmt"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/utility"
)

type DropsComponent struct {
	Items         map[string]int // item name to quantity
	DropChances   map[string]int // item name to drop chance (0-100)
	NumRolls      int            // number of times to roll for drops
	AlwaysDropAll bool           // if true, all items with be dropped
}

func (c *DropsComponent) GetType() ecs.ComponentType {
	return Drops
}

func (c *DropsComponent) GetDrops() map[string]int {
	if c.AlwaysDropAll {
		return c.Items
	}

	fmt.Println("Calculating drops for", c.Items)
	// TODO - Look more into this, but fine for POC.
	dropped := make(map[string]int)
	for item, quantity := range c.Items {
		chance := c.DropChances[item]
		for i := 0; i < c.NumRolls; i++ {
			if utility.GetRandom(0, 100) < chance {
				if quantity <= 0 {
					quantity = 1
				}
				dropped[item] += quantity
			}
		}
	}

	return dropped
}
