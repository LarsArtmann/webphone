package domain

import "time"

// FaxDirection tells which way a fax traveled.
type FaxDirection string

const (
	// FaxOutbound was sent by the extension.
	FaxOutbound FaxDirection = "out"
	// FaxInbound arrived at the extension.
	FaxInbound FaxDirection = "in"
)

// FaxStatus is the fax job state machine:
//
//	outbound: queued → sending → transmitted | failed
//	inbound:  received
//
// "failed" is terminal; "sending" advances only via gateway status updates.
type FaxStatus string

const (
	// FaxQueued: PDF stored locally, gateway not yet called.
	FaxQueued FaxStatus = "queued"
	// FaxSending: gateway accepted the job; waiting for the verdict.
	FaxSending FaxStatus = "sending"
	// FaxTransmitted: the far end confirmed all pages.
	FaxTransmitted FaxStatus = "transmitted"
	// FaxReceived: an inbound fax document landed in the inbox.
	FaxReceived FaxStatus = "received"
	// FaxFailed: transmission failed; Error carries the reason.
	FaxFailed FaxStatus = "failed"
)

// FaxJob is one fax transmission attempt (outbound) or one received
// document (inbound). The PDF lives on disk at DocumentPath.
type FaxJob struct {
	ID           FaxID
	Owner        Extension
	Remote       Phone
	Direction    FaxDirection
	Status       FaxStatus
	Pages        int
	DocumentPath string
	ProviderRef  string
	Error        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// InboundFax is a received fax document crossing the gateway boundary.
type InboundFax struct {
	Owner     Extension
	From      Phone
	Pages     int
	PDFBytes  []byte
	Received  time.Time
	ProviderRef string
}
