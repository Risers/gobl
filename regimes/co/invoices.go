package co

import (
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/currency"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/common"
	"github.com/invopop/gobl/tax"
	"github.com/invopop/validation"
)

var invoiceTags = []*tax.KeyDefinition{
	// Simplified invoices when a client ID is not available. For conversion to local formats,
	// the correct client ID data will need to be provided automatically.
	{
		Key: common.TagSimplified,
		Name: i18n.String{
			i18n.EN: "Simplified Invoice",
			i18n.ES: "Factura Simplificada",
		},
		Desc: i18n.String{
			i18n.EN: "Used for B2C transactions when the client details are not available, check with local authorities for limits.",
			i18n.ES: "Usado para transacciones B2C cuando los detalles del cliente no están disponibles, consulte con las autoridades locales para los límites.",
		},
	},
}

type invoiceValidator struct {
	inv *bill.Invoice
}

func validateInvoice(inv *bill.Invoice) error {
	v := &invoiceValidator{inv: inv}
	return v.validate()
}

func (v *invoiceValidator) validate() error {
	inv := v.inv
	return validation.ValidateStruct(inv,
		validation.Field(&inv.Currency, validation.In(currency.COP)),
		validation.Field(&inv.Type,
			validation.In(
				bill.InvoiceTypeStandard,
				bill.InvoiceTypeCreditNote,
				bill.InvoiceTypeProforma,
			),
		),
		validation.Field(&inv.Supplier,
			validation.Required,
			validation.By(v.validParty),
			validation.By(v.validSupplier),
		),
		validation.Field(&inv.Customer,
			validation.When(
				!inv.Tax.ContainsTag(common.TagSimplified),
				validation.Required,
			),
			validation.By(v.validParty),
		),
		validation.Field(&inv.Preceding,
			validation.When(
				inv.Type.In(bill.InvoiceTypeCreditNote),
				validation.Required,
			),
			validation.Each(validation.By(v.preceding))),
		validation.Field(&inv.Outlays, validation.Empty),
	)
}

func (v *invoiceValidator) validParty(value interface{}) error {
	obj, _ := value.(*org.Party)
	if obj == nil || obj.TaxID == nil {
		return nil
	}
	return validation.ValidateStruct(obj,
		validation.Field(&obj.TaxID,
			validation.Required,
			validation.When(
				obj.TaxID.Country.In(l10n.CO),
				tax.RequireIdentityCode,
				validation.By(v.validTaxIdentity),
			),
		),
		validation.Field(&obj.Addresses,
			validation.When(
				obj.TaxID.Country.In(l10n.CO),
				validation.Length(1, 0),
			),
		),
	)
}

func (v *invoiceValidator) validSupplier(value interface{}) error {
	obj, _ := value.(*org.Party)
	if obj == nil || obj.TaxID == nil {
		return nil
	}
	return validation.ValidateStruct(obj,
		validation.Field(&obj.TaxID,
			tax.IdentityTypeIn(TaxIdentityTypeTIN),
		),
	)
}

func (v *invoiceValidator) validTaxIdentity(value interface{}) error {
	obj, _ := value.(*tax.Identity)
	if obj == nil {
		return nil
	}
	return validation.ValidateStruct(obj,
		validation.Field(&obj.Zone,
			tax.ZoneIn(zones),
			validation.When(
				obj.Type.In(TaxIdentityTypeTIN),
				validation.Required,
			),
		),
	)
}

func (v *invoiceValidator) preceding(value interface{}) error {
	obj, ok := value.(*bill.Preceding)
	if !ok || obj == nil {
		return nil
	}
	return validation.ValidateStruct(obj,
		validation.Field(&obj.CorrectionMethod, validation.Required, isValidCorrectionMethodKey),
		validation.Field(&obj.Reason, validation.Required),
	)
}
