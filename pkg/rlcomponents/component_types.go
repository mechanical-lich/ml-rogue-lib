package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// Component type constants (Position and Direction are defined in their own files).
const (
	Health          ecs.ComponentType = "Health"
	Stats           ecs.ComponentType = "Stats"
	Initiative      ecs.ComponentType = "Initiative"
	MyTurn          ecs.ComponentType = "MyTurn"
	Dead            ecs.ComponentType = "Dead"
	Description     ecs.ComponentType = "Description"
	Solid           ecs.ComponentType = "Solid"
	Inanimate       ecs.ComponentType = "Inanimate"
	NeverSleep      ecs.ComponentType = "NeverSleep"
	Nocturnal       ecs.ComponentType = "Nocturnal"
	WanderAI        ecs.ComponentType = "WanderAI"
	HostileAI       ecs.ComponentType = "HostileAI"
	DefensiveAI     ecs.ComponentType = "DefensiveAI"
	AIMemory        ecs.ComponentType = "AIMemory"
	Alerted         ecs.ComponentType = "Alerted"
	Poisoned        ecs.ComponentType = "Poisoned"
	Poisonous       ecs.ComponentType = "Poisonous"
	Burning         ecs.ComponentType = "Burning"
	Regeneration    ecs.ComponentType = "Regeneration"
	LightSensitive  ecs.ComponentType = "LightSensitive"
	Light           ecs.ComponentType = "Light"
	Door            ecs.ComponentType = "Door"
	Food            ecs.ComponentType = "Food"
	Inventory       ecs.ComponentType = "Inventory"
	Item            ecs.ComponentType = "Item"
	Armor           ecs.ComponentType = "Armor"
	Weapon          ecs.ComponentType = "Weapon"
	AsciiAppearance ecs.ComponentType = "AsciiAppearance"
	Interaction     ecs.ComponentType = "Interaction"
	Key             ecs.ComponentType = "Key"
	TurnTaken       ecs.ComponentType = "TurnTaken"
	Size            ecs.ComponentType = "Size"
	Body            ecs.ComponentType = "Body"
	BodyInventory   ecs.ComponentType = "BodyInventory"
	Drops           ecs.ComponentType = "Drops"
	EnergyType      ecs.ComponentType = "Energy"
)

func init() {
	ecs.InanimateComponentType = Inanimate
}
