package mx

import (
	"errors"

	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/pay"
	"github.com/invopop/validation"
)

type invoiceValidator struct {
	inv *bill.Invoice
}

func validateInvoice(inv *bill.Invoice) error {
	v := &invoiceValidator{inv: inv}
	return v.validate()
}

func (v *invoiceValidator) validate() error {
	if err := v.validateScenarios(); err != nil {
		return err
	}

	inv := v.inv
	return validation.ValidateStruct(inv,
		validation.Field(&inv.Customer,
			validation.Required,
			validation.By(v.validCustomer),
		),
		validation.Field(&inv.Lines,
			validation.Each(
				validation.By(v.validLine),
				validation.Skip, // Prevents each line's `ValidateWithContext` function from being called again.
			),
			validation.Skip, // Prevents each line's `ValidateWithContext` function from being called again.
		),
		validation.Field(&inv.Payment,
			validation.Required,
			validation.By(v.validPayment),
		),
	)
}

func (v *invoiceValidator) validateScenarios() error {
	ss := v.inv.ScenarioSummary()

	if ss.Codes[KeySATUsoCFDI] == "" {
		return errors.New("tax tags are missing or don’t map to a UsoCFDI code")
	}

	return nil
}

func (v *invoiceValidator) validCustomer(value interface{}) error {
	obj, _ := value.(*org.Party)
	if obj == nil {
		return nil
	}
	return validation.ValidateStruct(obj,
		validation.Field(&obj.TaxID, validation.Required),
	)
}

func (v *invoiceValidator) validLine(value interface{}) error {
	line, _ := value.(*bill.Line)
	if line == nil {
		return nil
	}

	return validation.ValidateStruct(line,
		validation.Field(&line.Quantity, num.Positive),
		validation.Field(&line.Total, num.Positive),
		validation.Field(&line.Taxes,
			validation.Required,
			validation.Skip, // Prevents each tax's `ValidateWithContext` function from being called again.
		),
	)
}

func (v *invoiceValidator) validPayment(value interface{}) error {
	pay, _ := value.(*bill.Payment)
	if pay == nil {
		return nil
	}
	return validation.ValidateStruct(pay,
		validation.Field(&pay.Instructions,
			validation.Required,
			validation.By(v.validatePayInstructions),
		),
	)
}

func (v *invoiceValidator) validatePayInstructions(value interface{}) error {
	instr, _ := value.(*pay.Instructions)
	if instr == nil {
		return nil
	}

	return validation.ValidateStruct(instr,
		validation.Field(&instr.Key, isValidPaymentMeanKey),
	)
}

var isValidPaymentMeanKey = validation.In(validPaymentMeanKeys()...)

func validPaymentMeanKeys() []interface{} {
	keys := make([]interface{}, len(paymentMeansKeyDefinitions))
	for i, keyDef := range paymentMeansKeyDefinitions {
		keys[i] = keyDef.Key
	}

	return keys
}
