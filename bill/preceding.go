package bill

import (
	"encoding/json"

	"github.com/invopop/gobl/cal"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/head"
	"github.com/invopop/gobl/uuid"
	"github.com/invopop/validation"
)

// Preceding allows for information to be provided about a previous invoice that this one
// will replace, subtract from, or add to. If this is used, the invoice type code will most likely need
// to be set to `corrective`, `credit-note`, or similar.
type Preceding struct {
	// Preceding document's UUID if available can be useful for tracing.
	UUID *uuid.UUID `json:"uuid,omitempty" jsonschema:"title=UUID"`
	// Series identification code
	Series string `json:"series,omitempty" jsonschema:"title=Series"`
	// Code of the previous document.
	Code string `json:"code" jsonschema:"title=Code"`
	// The issue date of the previous document.
	IssueDate *cal.Date `json:"issue_date,omitempty" jsonschema:"title=Issue Date"`
	// Human readable description on why the preceding invoice is being replaced.
	Reason string `json:"reason,omitempty" jsonschema:"title=Reason"`
	// Seals of approval from other organisations that may need to be listed.
	Stamps []*head.Stamp `json:"stamps,omitempty" jsonschema:"title=Stamps"`
	// Tax regime specific key reflecting the method used to correct the preceding invoice.
	CorrectionMethod cbc.Key `json:"correction_method,omitempty" jsonschema:"title=Correction Method"`
	// Tax regime specific keys reflecting what has been changed from the previous invoice.
	Changes []cbc.Key `json:"changes,omitempty" jsonschema:"title=Changes"`
	// Tax period in which the previous invoice had an effect required by some tax regimes and formats.
	Period *cal.Period `json:"period,omitempty" jsonschema:"title=Period"`
	// Additional semi-structured data that may be useful in specific regions
	Meta cbc.Meta `json:"meta,omitempty" jsonschema:"title=Meta"`
}

// UnmarshalJSON is used to handle the refactor away from "corrections" to "changes"
func (p *Preceding) UnmarshalJSON(data []byte) error {
	type Alias Preceding
	aux := &struct {
		Corrections []cbc.Key `json:"corrections,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(p),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Corrections) != 0 {
		p.Changes = aux.Corrections
	}
	return nil
}

// Validate ensures the preceding details look okay
func (p *Preceding) Validate() error {
	return validation.ValidateStruct(p,
		validation.Field(&p.UUID),
		validation.Field(&p.Series),
		validation.Field(&p.Code, validation.Required),
		validation.Field(&p.IssueDate, cal.DateNotZero()),
		validation.Field(&p.Stamps),
		validation.Field(&p.CorrectionMethod),
		validation.Field(&p.Changes),
		validation.Field(&p.Period),
		validation.Field(&p.Meta),
	)
}
