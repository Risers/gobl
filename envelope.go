package gobl

import (
	"context"
	"errors"
	"fmt"

	"github.com/invopop/validation"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/internal"
	"github.com/invopop/gobl/schema"
	"github.com/invopop/gobl/uuid"
)

// Envelope wraps around a document adding headers and
// digital signatures. An Envelope is similar to a regular envelope
// in the physical world, it keeps the contents safe and helps
// get the document where its needed.
type Envelope struct {
	// Schema identifies the schema that should be used to understand this document
	Schema schema.ID `json:"$schema" jsonschema:"title=JSON Schema ID"`
	// Details on what the contents are
	Head *Header `json:"head" jsonschema:"title=Header"`
	// The data inside the envelope
	Document *Document `json:"doc" jsonschema:"title=Document"`
	// JSON Web Signatures of the header
	Signatures []*dsig.Signature `json:"sigs,omitempty" jsonschema:"title=Signatures"`
}

// EnvelopeSchema sets the general definition of the schema ID for this version of the
// envelope.
var EnvelopeSchema = schema.GOBL.Add("envelope")

// NewEnvelope builds a new envelope object ready for data to be inserted
// and signed. If you are loading data from json, you can safely use a regular
// `new(Envelope)` call directly.
func NewEnvelope() *Envelope {
	e := new(Envelope)
	e.Schema = EnvelopeSchema
	e.Head = NewHeader()
	e.Document = new(Document)
	e.Signatures = make([]*dsig.Signature, 0)
	return e
}

// Envelop is a convenience method that will build a new envelope and insert
// the contents document provided in a single swoop. The resulting envelope
// will still need to be signed afterwards.
func Envelop(doc interface{}) (*Envelope, error) {
	e := NewEnvelope()
	if err := e.Insert(doc); err != nil {
		return nil, err
	}
	return e, nil
}

// Validate ensures that the envelope contains everything it should to be considered valid GoBL.
func (e *Envelope) Validate() error {
	return e.ValidateWithContext(context.Background())
}

// ValidateWithContext ensures that the envelope contains everything it should to be considered valid GoBL.
func (e *Envelope) ValidateWithContext(ctx context.Context) error {
	ctx = context.WithValue(ctx, internal.KeyDraft, e.Head != nil && e.Head.Draft)
	err := validation.ValidateStructWithContext(ctx, e,
		validation.Field(&e.Schema, validation.Required),
		validation.Field(&e.Head, validation.Required),
		validation.Field(&e.Document, validation.Required), // this will also check payload
		validation.Field(&e.Signatures, validation.When(e.Head == nil || e.Head.Draft, validation.Empty)),
	)
	if err != nil {
		return err
	}
	return e.verifyDigest()
}

func (e *Envelope) verifyDigest() error {
	d1 := e.Head.Digest
	d2, err := e.Document.Digest()
	if err != nil {
		return err
	}
	if err := d1.Equals(d2); err != nil {
		return fmt.Errorf("document: %w", err)
	}
	return nil
}

// Sign uses the private key to sign the envelope headers. The header
// draft flag will be set to false and validation is performed so that
// only valid non-draft documents will be signed.
func (e *Envelope) Sign(key *dsig.PrivateKey) error {
	if e.Head == nil {
		return ErrValidation.WithCause(errors.New("missing header"))
	}
	e.Head.Draft = false
	if err := e.Validate(); err != nil {
		return ErrValidation.WithCause(err)
	}
	sig, err := key.Sign(e.Head)
	if err != nil {
		return ErrSignature.WithCause(err)
	}
	e.Signatures = append(e.Signatures, sig)
	return nil
}

// Insert takes the provided document and inserts it into this
// envelope. Normalize will be called automatically.
func (e *Envelope) Insert(doc interface{}) error {
	if e.Head == nil {
		return ErrInternal.WithErrorf("missing head")
	}
	if doc == nil {
		return ErrNoDocument
	}

	if d, ok := doc.(*Document); ok {
		e.Document = d
	} else {
		var err error
		e.Document, err = NewDocument(doc)
		if err != nil {
			return err
		}
	}

	if err := e.normalize(); err != nil {
		return err
	}

	return nil
}

// Calculate will perform any normalization and calculation methods on the
// envelope and document to ensure the basic data is correct.
//
// Deprecated: use the Normalize method instead.
func (e *Envelope) Calculate() error {
	return e.Normalize()
}

// Normalize is used to normalize the envelope and document data alongside
// any calculations required to complete the data.
// Headers will also be refreshed to ensure they have the latest valid
// digest data.
func (e *Envelope) Normalize() error {
	if e.Document == nil {
		return ErrNoDocument
	}
	if e.Document.IsEmpty() {
		return ErrNoDocument
	}

	return e.normalize()
}

func (e *Envelope) normalize() error {
	// Always set our schema version
	e.Schema = EnvelopeSchema

	// arm doors and cross check
	if err := e.Document.Normalize(); err != nil {
		return ErrCalculation.WithCause(err)
	}

	// Double check the header looks okay
	if e.Head == nil {
		e.Head = NewHeader()
	}
	if e.Head.UUID.IsZero() {
		e.Head.UUID = uuid.MakeV1()
	}
	var err error
	e.Head.Digest, err = e.Document.Digest()
	if err != nil {
		return err
	}

	return nil
}

// Extract the contents of the envelope into the provided document type.
func (e *Envelope) Extract() interface{} {
	if e.Document == nil {
		return nil
	}
	return e.Document.Instance()
}

// Correct will attempt to build a new envelope as a correction of the
// current envelope contents, if possible.
func (e *Envelope) Correct(opts ...cbc.Option) (*Envelope, error) {
	// Determine any extra options
	switch e.Document.Instance().(type) {
	case *bill.Invoice:
		// Special case for invoices so that we copy over
		// the stamps from the original invoice headers.
		if len(e.Head.Stamps) > 0 {
			opts = append(opts, bill.WithStamps(e.Head.Stamps))
		}
	}

	nd, err := e.Document.Clone()
	if err != nil {
		return nil, err
	}
	if err := nd.Correct(opts...); err != nil {
		return nil, err
	}

	// Create a completely new envelope with a new set of data.
	return Envelop(nd)
}
