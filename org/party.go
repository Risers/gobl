package org

import (
	"context"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/tax"
	"github.com/invopop/gobl/uuid"

	"github.com/invopop/validation"
)

// Party represents a person or business entity.
type Party struct {
	// Unique identity code
	UUID *uuid.UUID `json:"uuid,omitempty" jsonschema:"title=UUID"`
	// Legal name or representation of the organization.
	Name string `json:"name" jsonschema:"title=Name"`
	// Alternate short name.
	Alias string `json:"alias,omitempty" jsonschema:"title=Alias"`
	// The entity's legal ID code used for tax purposes. They may have other numbers, but we're only interested in those valid for tax purposes.
	TaxID *tax.Identity `json:"tax_id,omitempty" jsonschema:"title=Tax Identity"`
	// Set of codes used to identify the party in other systems.
	Identities []*Identity `json:"identities,omitempty" jsonschema:"title=Identities"`
	// Details of physical people who represent the party.
	People []*Person `json:"people,omitempty" jsonschema:"title=People"`
	// Digital inboxes used for forwarding electronic versions of documents
	Inboxes []*Inbox `json:"inboxes,omitempty" jsonschema:"title=Inboxes"`
	// Regular post addresses for where information should be sent if needed.
	Addresses []*Address `json:"addresses,omitempty" jsonschema:"title=Postal Addresses"`
	// Electronic mail addresses
	Emails []*Email `json:"emails,omitempty" jsonschema:"title=Email Addresses"`
	// Public websites that provide further information about the party.
	Websites []*Website `json:"websites,omitempty" jsonschema:"title=Websites"`
	// Regular telephone numbers
	Telephones []*Telephone `json:"telephones,omitempty" jsonschema:"title=Telephone Numbers"`
	// Additional registration details about the company that may need to be included in a document.
	Registration *Registration `json:"registration,omitempty" jsonschema:"title=Registration"`
	// Images that can be used to identify the party visually.
	Logos []*Image `json:"logos,omitempty" jsonschema:"title=Logos"`
	// Any additional semi-structured information that does not fit into the rest of the party.
	Meta cbc.Meta `json:"meta,omitempty" jsonschema:"title=Meta"`
}

// Normalize performs any calculations required on the Party or
// it's properties, like the tax identity.
func (p *Party) Normalize() error {
	if p.TaxID == nil {
		return nil
	}
	if err := p.TaxID.Normalize(); err != nil {
		return err
	}
	r := p.TaxID.Regime()
	if r == nil {
		return nil // nothing to do here
	}
	return r.NormalizeObject(p)
}

// Validate is used to check the party's data meets minimum expectations.
func (p *Party) Validate() error {
	return p.ValidateWithContext(context.Background())
}

// ValidateWithContext is used to check the party's data meets minimum expectations.
func (p *Party) ValidateWithContext(ctx context.Context) error {
	return tax.ValidateStructWithRegime(ctx, p,
		validation.Field(&p.Name, validation.Required),
		validation.Field(&p.TaxID),
		validation.Field(&p.Identities),
		validation.Field(&p.People),
		validation.Field(&p.Inboxes),
		validation.Field(&p.Addresses),
		validation.Field(&p.Emails),
		validation.Field(&p.Websites),
		validation.Field(&p.Telephones),
		validation.Field(&p.Registration),
		validation.Field(&p.Logos),
		validation.Field(&p.Meta),
	)
}
