package nl

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/invopop/validation"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/regimes/common"
	"github.com/invopop/gobl/tax"
)

const (
	vatLen = 12
)

var errInvalidVAT = errors.New("invalid VAT number")

// ValidateTaxIdentity looks at the provided code, determines the type, and performs
// the calculations required to determine if it is valid.
// These methods assume the code has already been cleaned and only
// contains upper-case letters and numbers.
func ValidateTaxIdentity(tID *tax.Identity) error {
	return validation.ValidateStruct(tID,
		validation.Field(&tID.Code, validation.By(validateTaxCode)),
	)
}

func validateTaxCode(value interface{}) error {
	code, ok := value.(cbc.Code)
	if !ok {
		return nil
	}
	if code == "" {
		return nil
	}
	if len(code) != vatLen {
		return errors.New("invalid length")
	}
	if code[9] != 'B' {
		return errors.New("invalid company code")
	}
	return validateDigits(code[0:9], code[10:12])
}

// NormalizeTaxIdentity removes any whitespace or separation characters and ensures all letters are
// uppercase. It'll also remove the "NL" part at beginning if present such as required
// for EU VIES system which is redundant and not used in the validation process.
func NormalizeTaxIdentity(tID *tax.Identity) error {
	return common.NormalizeTaxIdentity(tID)
}

func validateDigits(code, check cbc.Code) error {
	num, err := strconv.ParseInt(string(code), 10, 64)
	if err != nil {
		return errInvalidVAT
	}
	_, err = strconv.Atoi(string(check))
	if err != nil {
		return errInvalidVAT
	}

	ck := num % 10 // last part of code
	sum := mod11(num)

	// changes in 2020 mean that NL VAT numbers have a different check
	// digit and should be checked with Mod 97 (like an IBAN).
	if sum != ck && !checkMod97(fmt.Sprintf("NL%sB%s", code, check)) {
		return errors.New("checksum mismatch")
	}

	return nil
}

func mod11(num int64) int64 {
	var sum int64
	for i := 0; i < 8; i++ {
		num /= 10
		mul := int64(i) + 2
		sum += (num % 10) * mul
	}
	sum = sum % 11
	if sum > 9 {
		sum = 0
	}
	return sum
}

func checkMod97(code string) bool {
	// Convert ASCII numbers and letters to integers
	set := make([]int, len(code))
	for i, char := range code {
		if char >= 48 && char <= 57 { // 0 -- 9
			set[i] = int(char - 48)
		} else { // assume letters
			set[i] = int(char - 55)
		}
	}

	// Concatenate all the numbers into a single integer
	var r int
	for _, c := range set {
		r = r * 10
		if c > 9 { // only support up to 2 digits!
			r = r * 10
		}
		r = r + c
	}

	return (r % 97) == 1
}
