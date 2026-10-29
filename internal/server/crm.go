package server

import (
	"github.com/larsartmann/webphone/internal/crm"
	"github.com/larsartmann/webphone/internal/web/views"
)

// crmContacts is the ONE conversion from the integration seam's type to the
// view projection: the published name always, a deep link only when the
// deployment configured a template. Views never import internal/crm, so no
// surface can grow a second, disagreeing notion of "who this is".
//
// A nil resolver (test compositions, integration off) yields nil, which is
// what displayName's raw-number fallback already expects.
func crmContacts(people map[string]crm.Person, resolver *crm.Resolver) views.CRMContacts {
	if len(people) == 0 {
		return nil
	}

	contacts := make(views.CRMContacts, len(people))
	for number, person := range people {
		contacts[number] = views.CRMContact{
			Name: person.Name,
			URL:  resolver.ProfileURL(person.Ref),
		}
	}

	return contacts
}
