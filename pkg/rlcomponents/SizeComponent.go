package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

type SizeComponent struct {
	Width  int
	Height int
}

func (sc SizeComponent) GetType() ecs.ComponentType {
	return Size
}
