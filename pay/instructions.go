package pay

import (
	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/org"
	"github.com/invopop/jsonschema"
	"github.com/invopop/validation"
	"github.com/invopop/validation/is"
)

// Standard payment method codes. This is a heavily reduced list of practical
// codes which can be linked to UNTDID 4461 counterparts.
// If you require more payment method options, please send your pull requests.
const (
	MethodKeyAny            cbc.Key = "any" // Use any method available.
	MethodKeyCard           cbc.Key = "card"
	MethodKeyCreditTransfer cbc.Key = "credit-transfer"
	MethodKeyDebitTransfer  cbc.Key = "debit-transfer"
	MethodKeyCash           cbc.Key = "cash"
	MethodKeyCheque         cbc.Key = "cheque"
	MethodKeyCredit         cbc.Key = "credit"
	MethodKeyBankDraft      cbc.Key = "bank-draft"
	MethodKeyDirectDebit    cbc.Key = "direct-debit" // aka. Mandate
	MethodKeyOnline         cbc.Key = "online"       // Website from which payment can be made
	MethodKeyVoucher        cbc.Key = "voucher"
)

// MethodKeyDef is used to define each of the Method Keys
// that can be accepted by GOBL.
type MethodKeyDef struct {
	// Key being described
	Key cbc.Key `json:"key" jsonschema:"title=Key"`
	// Human value of the key
	Title string `json:"title" jsonschema:"title=Title"`
	// Details about the meaning of the key
	Description string `json:"description" jsonschema:"title=Description"`
	// UNTDID 4461 Equivalent Code
	UNTDID4461 cbc.Code `json:"untdid4461" jsonschema:"title=UNTDID 4461 Code"`
}

// MethodKeyDefinitions includes all the payment method keys that
// are accepted by GOBL.
var MethodKeyDefinitions = []MethodKeyDef{
	{MethodKeyAny, "Any", "Any method available, no preference.", "1"},                            // Instrument not defined
	{MethodKeyCard, "Card", "Credit or debit card.", "48"},                                        // Bank card
	{MethodKeyCreditTransfer, "Credit Transfer", "Sender initiated bank or wire transfer.", "30"}, // credit transfer
	{MethodKeyDebitTransfer, "Debit Transfer", "Receiver initiated bank or wire transfer.", "31"}, // debit transfer
	{MethodKeyCash, "Cash", "Cash in hand.", "10"},                                                // in cash
	{MethodKeyCheque, "Cheque", "Cheque from bank.", ""},                                          // cheque
	{MethodKeyCredit, "Credit", "Using credit from previous transactions with the supplier.", ""}, // credit
	{MethodKeyBankDraft, "Draft", "Bankers Draft or Bank Cheque.", ""},                            // Banker's draft,
	{MethodKeyDirectDebit, "Direct Debit", "Direct debit from the customers bank account.", "49"}, // direct debit
	{MethodKeyOnline, "Online", "Online or web payment.", "68"},                                   // online payment service
	{MethodKeyVoucher, "Voucher", "Gift voucher or coupon.", ""},                                  // voucher
}

// Instructions determine how the payment has or should be made. A
// single "key" exists in which the preferred payment method
// should be provided, all other details serve as a reference.
type Instructions struct {
	// How payment is expected or has been arranged to be collected
	Key cbc.Key `json:"key" jsonschema:"title=Key"`
	// Optional text description of the payment method
	Detail string `json:"detail,omitempty" jsonschema:"title=Detail"`
	// Remittance information, a text value used to link the payment with the invoice.
	Ref string `json:"ref,omitempty" jsonschema:"title=Ref"`
	// Instructions for sending payment via a bank transfer.
	CreditTransfer []*CreditTransfer `json:"credit_transfer,omitempty" jsonschema:"title=Credit Transfer"`
	// Details of the payment that will be made via a credit or debit card.
	Card *Card `json:"card,omitempty" jsonschema:"title=Card"`
	// A group of terms that can be used by the customer or payer to consolidate direct debit payments.
	DirectDebit *DirectDebit `json:"direct_debit,omitempty" jsonschema:"title=Direct Debit"`
	// Array of online payment options
	Online []*Online `json:"online,omitempty" jsonschema:"title=Online"`
	// Any additional instructions that may be required to make the payment.
	Notes string `json:"notes,omitempty" jsonschema:"title=Notes"`
	// Non-structured additional data that may be useful.
	Meta cbc.Meta `json:"meta,omitempty" jsonschema:"title=Meta"`
}

// Card contains simplified card holder data as a reference for the customer.
type Card struct {
	// Last 4 digits of the card's Primary Account Number (PAN).
	Last4 string `json:"last4" jsonschema:"title=Last 4"`
	// Name of the person whom the card belongs to.
	Holder string `json:"holder" jsonschema:"title=Holder Name"`
}

// DirectDebit defines the data that will be used to make the direct debit.
type DirectDebit struct {
	// Unique identifier assigned by the payee for referencing the direct debit.
	Ref string `json:"ref,omitempty" jsonschema:"title=Mandate Reference"`
	// Unique banking reference that identifies the payee or seller assigned by the bank.
	Creditor string `json:"creditor,omitempty" jsonschema:"title=Creditor ID"`
	// Account identifier to be debited by the direct debit.
	Account string `json:"account,omitempty" jsonschema:"title=Account"`
}

// CreditTransfer contains fields that can be used for making payments via
// a bank transfer or wire.
type CreditTransfer struct {
	// International Bank Account Number
	IBAN string `json:"iban,omitempty" jsonschema:"title=IBAN"`
	// Bank Identifier Code used for international transfers.
	BIC string `json:"bic,omitempty" jsonschema:"title=BIC"`
	// Account number, if IBAN not available.
	Number string `json:"number,omitempty" jsonschema:"title=Number"`
	// Name of the bank.
	Name string `json:"name,omitempty" jsonschema:"title=Name"`
	// Bank office branch address, not normally required.
	Branch *org.Address `json:"branch,omitempty" jsonschema:"title=Branch"`
}

// Online provides the details required to make a payment online using a website
type Online struct {
	// Descriptive name given to the online provider.
	Name string `json:"name,omitempty" jsonschema:"title=Name"`
	// Full URL to be used for payment.
	Address string `json:"addr" jsonschema:"title=Address"`
}

// UNTDID4461 provides the standard UNTDID 4461 code for the instruction's key.
func (i *Instructions) UNTDID4461() cbc.Code {
	for _, v := range MethodKeyDefinitions {
		if v.Key == i.Key {
			return v.UNTDID4461
		}
	}
	return cbc.CodeEmpty
}

// Validate ensures the Online method details look correct.
func (u *Online) Validate() error {
	return validation.ValidateStruct(u,
		validation.Field(&u.Address, validation.Required, is.URL),
	)
}

// Validate ensures the fields provided in the instructions are valid.
func (i *Instructions) Validate() error {
	return validation.ValidateStruct(i,
		validation.Field(&i.Key, validation.Required, isValidMethodKey),
		validation.Field(&i.CreditTransfer),
		validation.Field(&i.DirectDebit),
		validation.Field(&i.Online),
	)
}

var isValidMethodKey = validation.In(validMethodKeys()...)

func validMethodKeys() []interface{} {
	list := make([]interface{}, len(MethodKeyDefinitions))
	for i, v := range MethodKeyDefinitions {
		list[i] = v.Key
	}
	return list
}

// JSONSchemaExtend adds the method key definitions to the schema.
func (Instructions) JSONSchemaExtend(schema *jsonschema.Schema) {
	val, _ := schema.Properties.Get("key")
	prop, ok := val.(*jsonschema.Schema)
	if ok {
		prop.OneOf = make([]*jsonschema.Schema, len(MethodKeyDefinitions))
		for i, v := range MethodKeyDefinitions {
			prop.OneOf[i] = &jsonschema.Schema{
				Const:       v.Key,
				Title:       v.Title,
				Description: v.Description,
			}
		}
	}
}
