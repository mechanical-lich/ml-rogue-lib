package rlasciiclient

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/transport"
)

// AsciiCodec implements transport.SnapshotCodec for ASCII clients.
//
// Server side (Encode): configurable via EncodeFunc.
// The default encoder encodes every entity that has both rlcomponents.Position
// and rlcomponents.AsciiAppearance, using the entity's memory address as its
// snapshot ID. This is correct for LocalTransport (single process). For TCP
// multiplayer, provide a custom EncodeFunc that assigns stable string IDs
// (e.g. from an EntityIDComponent).
//
// Client side (Decode): built-in. Applies AsciiAppearanceComponent and
// PositionComponent from each EntitySnapshot onto the AsciiWorld. If the
// snapshot carries a game-specific position type, set ExtractPos to translate
// it into (x, y, z) coordinates.
type AsciiCodec struct {
	// EncodeFunc is called by the server each tick. If nil, DefaultEncode is used.
	EncodeFunc func(tick uint64, entities []*ecs.Entity) *transport.Snapshot

	// ExtractPos extracts (x, y, z) from an entity snapshot's component map.
	// If nil, the codec looks for rlcomponents.PositionComponent under the
	// rlcomponents.Position key.
	ExtractPos func(comps map[ecs.ComponentType]transport.ComponentData) (x, y, z int, ok bool)
}

// Compile-time assertion.
var _ transport.SnapshotCodec = (*AsciiCodec)(nil)

// Encode satisfies transport.SnapshotCodec. Delegates to EncodeFunc or DefaultEncode.
func (c *AsciiCodec) Encode(tick uint64, entities []*ecs.Entity) *transport.Snapshot {
	if c.EncodeFunc != nil {
		return c.EncodeFunc(tick, entities)
	}
	return DefaultEncode(tick, entities)
}

// Decode satisfies transport.SnapshotCodec. Applies the snapshot to an *AsciiWorld.
func (c *AsciiCodec) Decode(snap *transport.Snapshot, world any) {
	w := world.(*AsciiWorld)
	alive := make(map[string]bool, len(snap.Entities))

	for _, es := range snap.Entities {
		alive[es.ID] = true
		e := w.FindOrCreate(es.ID, es.Blueprint)

		// Apply ASCII appearance.
		if raw, ok := es.Components[rlcomponents.AsciiAppearance]; ok {
			if ac, ok := raw.(*rlcomponents.AsciiAppearanceComponent); ok {
				e.AddComponent(ac)
			}
		}

		// Apply position: use custom extractor if provided, else read rlcomponents.Position.
		if c.ExtractPos != nil {
			if x, y, z, ok := c.ExtractPos(es.Components); ok {
				setPos(e, x, y, z)
			}
		} else {
			if raw, ok := es.Components[rlcomponents.Position]; ok {
				if pc, ok := raw.(*rlcomponents.PositionComponent); ok {
					setPos(e, pc.X, pc.Y, pc.Z)
				}
			}
		}
	}

	w.RemoveNotIn(alive)
}

// setPos writes (x,y,z) into the entity's PositionComponent, creating it if absent.
func setPos(e *ecs.Entity, x, y, z int) {
	if !e.HasComponent(rlcomponents.Position) {
		e.AddComponent(&rlcomponents.PositionComponent{X: x, Y: y, Z: z})
		return
	}
	pc := e.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	pc.X = x
	pc.Y = y
	pc.Z = z
}

// DefaultEncode encodes all entities that carry both rlcomponents.Position and
// rlcomponents.AsciiAppearance. The snapshot ID is the entity's memory address,
// which is stable within a single process (LocalTransport).
func DefaultEncode(tick uint64, entities []*ecs.Entity) *transport.Snapshot {
	snaps := make([]*transport.EntitySnapshot, 0, len(entities))
	for _, e := range entities {
		if !e.HasComponent(rlcomponents.AsciiAppearance) || !e.HasComponent(rlcomponents.Position) {
			continue
		}
		snaps = append(snaps, &transport.EntitySnapshot{
			ID:        fmt.Sprintf("%p", e),
			Blueprint: e.Blueprint,
			Components: map[ecs.ComponentType]transport.ComponentData{
				rlcomponents.AsciiAppearance: e.GetComponent(rlcomponents.AsciiAppearance),
				rlcomponents.Position:        e.GetComponent(rlcomponents.Position),
			},
		})
	}
	return transport.NewSnapshot(tick, snaps)
}
