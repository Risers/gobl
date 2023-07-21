// Package mx provides the Mexican tax regime.
package mx

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/common"
	"github.com/invopop/gobl/tax"
)

func init() {
	tax.RegisterRegime(New())
}

// Custom keys used typically in meta or codes information.
const (
	KeySATFormaPago cbc.Key = "sat-forma-pago" // for mapping to c_FormaPago’s codes
	KeySATUsoCFDI   cbc.Key = "sat-uso-cfdi"   // for mapping to c_UsoCFDI’s codes

	IdentityTypeSAT cbc.Code = "SAT" // for custom codes mapped from identities (e.g. c_ClaveProdServ’s codes)
)

// New provides the tax region definition
func New() *tax.Regime {
	return &tax.Regime{
		Country:  l10n.MX,
		Currency: currency.MXN,
		Name: i18n.String{
			i18n.EN: "Mexico",
			i18n.ES: "México",
		},
		PaymentMeansKeys: paymentMeansKeyDefinitions, // pay.go
		Validator:        Validate,
		Normalizer:       Normalize,
		Tags:             invoiceTags,   // scenarios.go
		Scenarios:        scenarios,     // scenarios.go
		Categories:       taxCategories, // categories.go
	}
}

// Validate validates a document against the tax regime.
func Validate(doc interface{}) error {
	switch obj := doc.(type) {
	case *bill.Invoice:
		return validateInvoice(obj)
	case *tax.Identity:
		return validateTaxIdentity(obj)
	case *org.Item:
		return validateItem(obj)
	}
	return nil
}

// Normalize performs regime specific calculations.
func Normalize(doc interface{}) error {
	switch obj := doc.(type) {
	case *tax.Identity:
		return common.NormalizeTaxIdentity(obj)
	case *org.Item:
		return normalizeItem(obj)
	}
	return nil
}
