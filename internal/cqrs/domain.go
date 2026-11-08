package cqrs

import (
	"github.com/larsartmann/go-cqrs-lite/system/v4"
)

// Domain merges DomainConfig shares (commands run in order, projections
// and declared events concatenate) — the composition root's one-call
// assembly of the aggregate family.
func Domain(decls ...system.DomainConfig) system.DomainConfig {
	merged := system.DomainConfig{}
	for _, d := range decls {
		if d.Commands != nil {
			prev := merged.Commands
			merged.Commands = func(sys *system.System) {
				if prev != nil {
					prev(sys)
				}
				d.Commands(sys)
			}
		}
		merged.Projections = append(merged.Projections, d.Projections...)
		merged.Events = append(merged.Events, d.Events...)
		merged.ProjectionHostOptions = append(merged.ProjectionHostOptions, d.ProjectionHostOptions...)
	}
	return merged
}
