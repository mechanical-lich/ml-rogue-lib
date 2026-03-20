package rlasciiclient

import "github.com/mechanical-lich/mlge/ecs"

// AsciiWorld is the client's non-authoritative entity store used for ASCII rendering.
// It holds a flat entity slice and an ID index for O(1) find-or-create.
type AsciiWorld struct {
	Entities []*ecs.Entity
	byID     map[string]*ecs.Entity
}

// NewAsciiWorld allocates an empty AsciiWorld.
func NewAsciiWorld() *AsciiWorld {
	return &AsciiWorld{
		byID: make(map[string]*ecs.Entity, 256),
	}
}

// FindOrCreate returns the entity with the given ID, creating it if it does not exist.
func (w *AsciiWorld) FindOrCreate(id, blueprint string) *ecs.Entity {
	if e, ok := w.byID[id]; ok {
		return e
	}
	e := &ecs.Entity{Blueprint: blueprint}
	w.Entities = append(w.Entities, e)
	w.byID[id] = e
	return e
}

// RemoveNotIn removes all entities whose ID is absent from alive.
// Call this at the end of each Decode pass to cull despawned entities.
func (w *AsciiWorld) RemoveNotIn(alive map[string]bool) {
	for id := range w.byID {
		if !alive[id] {
			delete(w.byID, id)
		}
	}
	// Rebuild slice from the surviving map entries.
	w.Entities = w.Entities[:0]
	for _, e := range w.byID {
		w.Entities = append(w.Entities, e)
	}
}
