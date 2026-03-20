package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// AsciiAppearanceComponent represents the ASCII appearance of an entity.
type AsciiAppearanceComponent struct {
	Character string
	R, G, B   uint8
}

func (pc AsciiAppearanceComponent) GetType() ecs.ComponentType {
	return AsciiAppearance
}
