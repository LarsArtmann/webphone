package cqrs

import (
	"github.com/larsartmann/go-cqrs-lite/metaengine/v4"
	"github.com/larsartmann/go-cqrs-lite/record/v4"

	"github.com/larsartmann/webphone/internal/domain"
	"github.com/larsartmann/webphone/internal/readmodels"
)

// dialWindowsCollection / dialOverridesCollection mirror the readmodels
// engine's collection names so operators grepping either store see one
// vocabulary.
const (
	dialWindowsCollection   = "dial_windows"
	dialOverridesCollection = "dial_overrides"

	// dialWindowsGlobalCeiling mirrors the readmodels declaration's
	// collection-global safety net.
	dialWindowsGlobalCeiling = 100_000
)

// dialWindowKeyInput is the key-extraction input for the journal
// projections (same shape the readmodels engine declares).
type dialWindowKeyInput struct{ Key domain.DialWindowKey }

// dialWindowsProjection folds DialAttempted facts into the SAME view
// shape the server reads today (readmodels.DialWindowView).
func dialWindowsProjection() metaengine.QueryDecl[dialWindowKeyInput, readmodels.DialWindowView] {
	return metaengine.Query[dialWindowKeyInput, readmodels.DialWindowView](dialWindowsCollection,
		metaengine.OnRecord(domain.DialAttempted{},
			func(_ record.Record, e domain.DialAttempted) (domain.DialWindowKey, readmodels.DialWindowView) {
				return e.Key, readmodels.DialWindowView{Key: e.Key, Owner: e.Owner, Remote: e.Remote, At: e.At.Unix()}
			}),
		metaengine.OnRecord(domain.DialWindowExpunged{}, metaengine.Remove[readmodels.DialWindowView]()),
		metaengine.FilterOnField[readmodels.DialWindowView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.DialWindowView]("at", false),
		metaengine.WithCap(dialWindowsGlobalCeiling, "at"),
		metaengine.WithCapPerGroup("owner", dialWindowsMaxSlots),
		metaengine.Volume(humanScaleVolume),
	)
}

// dialOverridesProjection folds the one-shot bypass slots.
func dialOverridesProjection() metaengine.QueryDecl[dialWindowKeyInput, readmodels.DialOverrideView] {
	return metaengine.Query[dialWindowKeyInput, readmodels.DialOverrideView](dialOverridesCollection,
		metaengine.OnRecord(domain.DialOverrideSaved{},
			func(_ record.Record, e domain.DialOverrideSaved) (domain.DialWindowKey, readmodels.DialOverrideView) {
				return e.Key, readmodels.DialOverrideView{Key: e.Key, Owner: e.Owner, Remote: e.Remote, Consumed: e.Consumed, At: e.At.Unix()}
			}),
		metaengine.OnRecord(domain.DialOverrideExpunged{}, metaengine.Remove[readmodels.DialOverrideView]()),
		metaengine.FilterOnField[readmodels.DialOverrideView]("owner", metaengine.FilterEq),
		metaengine.SortOnField[readmodels.DialOverrideView]("at", false),
		metaengine.WithCap(dialWindowsGlobalCeiling, "at"),
		metaengine.WithCapPerGroup("owner", dialOverridesMaxSlots),
		metaengine.Volume(humanScaleVolume),
	)
}
