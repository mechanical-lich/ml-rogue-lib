package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// MyTurnComponent is a marker added to an entity when it is allowed to act this frame.
type MyTurnComponent struct{}

func (pc *MyTurnComponent) GetType() ecs.ComponentType {
	return MyTurn
}

// Shared sentinel — MyTurnComponent is stateless so one instance suffices.
var myTurnSentinel = &MyTurnComponent{}

func GetMyTurn() *MyTurnComponent {
	return myTurnSentinel
}
