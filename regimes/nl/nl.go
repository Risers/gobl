// Package nl provides the Dutch region definition
package nl

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/regimes/common"
	"github.com/invopop/gobl/tax"
)

func init() {
	tax.RegisterRegime(New())
}

// New provides the Dutch region definition
func New() *tax.Regime {
	return &tax.Regime{
		Country:  l10n.NL,
		Currency: "EUR",
		Name: i18n.String{
			i18n.EN: "The Netherlands",
			i18n.NL: "Nederland",
		},
		Validator:  Validate,
		Normalizer: Normalize,
		Categories: []*tax.Category{
			//
			// VAT
			//
			{
				Code: common.TaxCategoryVAT,
				Name: i18n.String{
					i18n.EN: "VAT",
					i18n.NL: "BTW",
				},
				Desc: i18n.String{
					i18n.EN: "Value Added Tax",
					i18n.NL: "Belasting Toegevoegde Waarde",
				},
				Retained: false,
				Rates: []*tax.Rate{
					{
						Key: common.TaxRateZero,
						Name: i18n.String{
							i18n.EN: "Zero Rate",
							i18n.NL: `0%-tarief`,
						},
						Values: []*tax.RateValue{
							{
								Percent: num.MakePercentage(0, 3),
							},
						},
					},
					{
						Key: common.TaxRateStandard,
						Name: i18n.String{
							i18n.EN: "Standard Rate",
							i18n.NL: "Standaardtarief",
						},
						Values: []*tax.RateValue{
							{
								Percent: num.MakePercentage(210, 3),
							},
						},
					},
					{
						Key: common.TaxRateReduced,
						Name: i18n.String{
							i18n.EN: "Reduced Rate",
							i18n.NL: "Gereduceerd Tarief",
						},
						Values: []*tax.RateValue{
							{
								Percent: num.MakePercentage(90, 3),
							},
						},
					},
				},
			},
		},
	}

}

// Validate checks the document type and determines if it can be validated.
func Validate(doc interface{}) error {
	switch obj := doc.(type) {
	case *bill.Invoice:
		return validateInvoice(obj)
	}
	return nil
}

// Normalize performs region specific calculations on the document.
func Normalize(doc interface{}) error {
	switch obj := doc.(type) {
	case *tax.Identity:
		return NormalizeTaxIdentity(obj)
	}
	return nil
}
