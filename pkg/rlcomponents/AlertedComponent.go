package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// AlertedComponent is a decaying status that keeps nocturnal/diurnal entities awake after combat.
type AlertedComponent struct {
	Duration int
}

func (pc AlertedComponent) GetType() ecs.ComponentType {
	return Alerted
}

func (pc *AlertedComponent) Decay() bool {
	pc.Duration--
	return pc.Duration <= 0
}
