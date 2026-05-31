package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DescriptionComponent holds an entity's display name, faction, and optional
// narrative text used by the interaction and announcement systems.
type DescriptionComponent struct {
	Name            string
	Faction         string
	ID              string   // Optional unique identifier used by the interaction system to reference this entity.
	Tags            []string // Optional group tags (e.g. "airlock", "sector_b") for multi-target triggers.
	LongDescription string   // Shown when the player examines the entity.
	Species         string
	Classification  string

	PassOverDescription   []string // Randomly chosen when a configured entity passes over this one.
	DeathAnnouncements    []string // Randomly chosen when a watcher sees this entity die. Defaults to "<Name> has died."
	ExcuseMeAnnouncements []string // Randomly chosen when a friendly bumps into this entity and they swap.
}

// DisplayName returns the archetype name when Species or Classification are
// set, otherwise the instance Name. Capitalises each word of the species and
// classification so blueprints can store them lowercase.
func (d *DescriptionComponent) DisplayName() string {
	if d.Species == "" && d.Classification == "" {
		return d.Name
	}
	parts := []string{}
	if d.Species != "" {
		parts = append(parts, titleCaseWords(d.Species))
	}
	if d.Classification != "" {
		parts = append(parts, titleCaseWords(d.Classification))
	}
	out := parts[0]
	for _, p := range parts[1:] {
		out += " " + p
	}
	return out
}

// titleCaseWords capitalises each space-separated word of s. "rage elite" →
// "Rage Elite". Single words are handled the same way.
func titleCaseWords(s string) string {
	if s == "" {
		return ""
	}
	out := ""
	upper := true
	for _, r := range s {
		if r == ' ' || r == '_' {
			out += " "
			upper = true
			continue
		}
		if upper {
			if r >= 'a' && r <= 'z' {
				r -= 'a' - 'A'
			}
			upper = false
		}
		out += string(r)
	}
	return out
}

func (pc DescriptionComponent) GetType() ecs.ComponentType {
	return Description
}
