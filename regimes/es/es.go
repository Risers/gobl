// Package es provides tax regime support for Spain.
package es

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/tax"
)

func init() {
	tax.RegisterRegime(New())
}

// Local tax category definitions which are not considered standard.
const (
	TaxCategoryIRPF cbc.Code = "IRPF"
	TaxCategoryIGIC cbc.Code = "IGIC"
	TaxCategoryIPSI cbc.Code = "IPSI"
)

// Specific tax rate codes.
const (
	// IRPF non-standard Rates (usually for self-employed)
	TaxRatePro                cbc.Key = "pro"                 // Professional Services
	TaxRateProStart           cbc.Key = "pro-start"           // Professionals, first 2 years
	TaxRateModules            cbc.Key = "modules"             // Module system
	TaxRateAgriculture        cbc.Key = "agriculture"         // Agricultural
	TaxRateAgricultureSpecial cbc.Key = "agriculture-special" // Agricultural special
	TaxRateCapital            cbc.Key = "capital"             // Rental or Interest

	// Special tax rate surcharge extension
	TaxRateEquivalence cbc.Key = "eqs"
)

// Official stamps or codes validated by government agencies
const (
	// TicketBAI (Basque Country) codes used for stamps.
	StampProviderTBAICode cbc.Key = "tbai-code"
	StampProviderTBAIQR   cbc.Key = "tbai-qr"
)

// Inbox key and role definitions
const (
	InboxKeyFACE cbc.Key = "face"

	// Main roles defined in FACE
	InboxRoleFiscal    cbc.Key = "fiscal"    // Fiscal / 01
	InboxRoleRecipient cbc.Key = "recipient" // Receptor / 02
	InboxRolePayer     cbc.Key = "payer"     // Pagador / 03
	InboxRoleCustomer  cbc.Key = "customer"  // Comprador / 04

)

// Custom keys used typically in meta information.
const (
	KeyAddressCode                 cbc.Key = "post"
	KeyFacturaE                    cbc.Key = "facturae"
	KeyFacturaETaxTypeCode         cbc.Key = "facturae-tax-type-code"
	KeyFacturaEInvoiceDocumentType cbc.Key = "facturae-invoice-document-type"
	KeyFacturaEInvoiceClass        cbc.Key = "facturae-invoice-class"
	KeyTicketBAICausaExencion      cbc.Key = "ticketbai-causa-exencion"
	KeyTicketBAIIDType             cbc.Key = "ticketbai-id-type"
)

// New provides the Spanish tax regime definition
func New() *tax.Regime {
	return &tax.Regime{
		Country:  l10n.ES,
		Currency: "EUR",
		Name: i18n.String{
			i18n.EN: "Spain",
			i18n.ES: "España",
		},
		Zones:            zones,
		Tags:             invoiceTags,
		IdentityTypeKeys: taxIdentityTypeDefinitions,
		Categories:       taxCategories,
		ItemKeys:         itemKeyDefinitions, // items.go
		Validator:        Validate,
		Normalizer:       Normalize,
		Scenarios: []*tax.ScenarioSet{
			invoiceScenarios,
		},
		Preceding: &tax.PrecedingDefinitions{
			Types: []cbc.Key{
				bill.InvoiceTypeCorrective,
			},
			Corrections:       correctionList,
			CorrectionMethods: correctionMethodList,
		},
	}
}

// Validate checks the document type and determines if it can be validated.
func Validate(doc interface{}) error {
	switch obj := doc.(type) {
	case *bill.Invoice:
		return validateInvoice(obj)
	case *tax.Identity:
		return validateTaxIdentity(obj)
	}
	return nil
}

// Normalize will perform any regime specific calculations.
func Normalize(doc interface{}) error {
	switch obj := doc.(type) {
	case *tax.Identity:
		return normalizeTaxIdentity(obj)
	}
	return nil
}
