package crm

import (
	"testing"

	"github.com/larsartmann/webphone/internal/domain"
)

// TestCRMOuTcomeCoversTheWholeDispositionVocabulary pins the bridge in both
// directions of the risk: every webphone disposition maps to a value the CRM
// ACCEPTS (a miss here is a 400 on every call, which the island shows as a
// 502 toast), and nothing maps to a value outside the CRM's closed set.
func TestCRMOuTcomeCoversTheWholeDispositionVocabulary(t *testing.T) {
	t.Parallel()

	// The CRM's closed vocabulary (machineapi contact.ParseOutcome). Kept as
	// a literal copy on purpose: a consumer test must break when the
	// producer's set changes, not silently agree with itself.
	crmVocabulary := map[string]bool{
		"": true, "connected": true, "voicemail": true, "ivr": true,
		"no_answer": true, "wrong_number": true, "dnc": true, "disconnected": true,
	}

	dispositions := []domain.Disposition{
		domain.DispositionAnswered, domain.DispositionMissed, domain.DispositionVoicemail,
		domain.DispositionIVR, domain.DispositionFailed, domain.DispositionUnknown,
	}

	for _, disposition := range dispositions {
		mapped, ok := outcomeForDisposition[disposition]
		if !ok {
			t.Errorf("disposition %q has no CRM outcome mapping (the CRM would reject the raw value)", disposition)

			continue
		}

		if !crmVocabulary[mapped] {
			t.Errorf("disposition %q maps to %q, which the CRM's closed vocabulary does not accept", disposition, mapped)
		}
	}

	if len(outcomeForDisposition) != len(dispositions) {
		t.Errorf("mapping covers %d dispositions, the domain has %d: a new disposition needs a mapping row",
			len(outcomeForDisposition), len(dispositions))
	}

	// An unknown/future disposition degrades to "no outcome recorded" rather
	// than sending something the CRM rejects.
	if got := crmOutcome(domain.Disposition("brand-new")); got != "" {
		t.Errorf("unmapped disposition = %q, want the empty (unrecorded) outcome", got)
	}
}
