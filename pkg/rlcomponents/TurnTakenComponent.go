package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent is a marker added to an entity when it is allowed to act this frame.
type TurnTakenComponent struct{}

func (pc *TurnTakenComponent) GetType() ecs.ComponentType {
	return TurnTaken
}

// Shared sentinel — TurnTakenComponent is stateless so one instance suffices.
var turnTakenSentinel = &TurnTakenComponent{}

func GetTurnTaken() *TurnTakenComponent {
	return turnTakenSentinel
}
