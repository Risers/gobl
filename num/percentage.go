package num

import (
	"bytes"
	"fmt"

	"github.com/invopop/jsonschema"
)

var (
	factor1   = MakeAmount(1, 0)
	factor100 = MakeAmount(100, 0)
)

// Percentage wraps around the regular Amount handler to provide support
// for percentage values, especially useful for tax rates.
type Percentage struct {
	Amount
}

// NewPercentage provides a new pointer to a Percentage value.
// Using MakePercentage is recommend, but this is useful for handling
// nil values.
func NewPercentage(value int64, exp uint32) *Percentage {
	p := MakePercentage(value, exp)
	return &p
}

// MakePercentage will make a new Percentage instance with the provided
// value and exponent.
func MakePercentage(value int64, exp uint32) Percentage {
	return Percentage{Amount{value: value, exp: exp}}
}

// PercentageFromString builds a percentage value from a provided string.
// The % symbol will be removed automatically and rescale the stored amount
// to make future calculations easier. For example, the following two strings
// will be interpreted equally:
//
//   - `0.160`
//   - `16.0%`
func PercentageFromString(str string) (Percentage, error) {
	p := Percentage{}
	l := len(str)
	if l == 0 {
		return p, nil
	}
	rescale := false
	if str[l-1:] == "%" {
		str = str[:l-1]
		rescale = true
	}

	var err error
	p.Amount, err = AmountFromString(str)
	if err != nil {
		return p, err
	}
	if rescale {
		e := p.exp
		p.Amount = p.Amount.Rescale(e + 2).Divide(factor100)
	}

	return p, nil
}

// String outputs the percentage value in a human readable way including
// the percentage symbol.
func (p Percentage) String() string {
	return p.StringWithoutSymbol() + "%"
}

// StringWithoutSymbol provides the percent value without a percent symbol.
func (p Percentage) StringWithoutSymbol() string {
	e := int64(p.Amount.exp) - 2
	if e < 0 {
		e = 0
	}
	v := p.Amount.Multiply(factor100).Rescale(uint32(e))
	return v.String()
}

// Rescale will rescale the percentage value to the provided exponent.
func (p Percentage) Rescale(exp uint32) Percentage {
	return Percentage{Amount: p.Amount.Rescale(exp)}
}

// Of calculates the "percent of" the provided amount. The exponent of the
// provided amount is used.
func (p Percentage) Of(a Amount) Amount {
	return a.Multiply(p.Amount)
}

// From calculates what "percent from" the provided amount would result
// assuming the rate has already been applied.
func (p Percentage) From(a Amount) Amount {
	x := a.Divide(p.Factor())
	return a.Subtract(x)
}

// Factor returns the percentage amount as a factor, essentially
// adding 1 to the rate.
func (p Percentage) Factor() Amount {
	return p.Amount.Add(factor1)
}

// Equals wraps around the amount comparison to see if the two percentages
// have the same value.
func (p Percentage) Equals(p2 Percentage) bool {
	return p.Amount.Equals(p2.Amount)
}

// MarshalText provides the byte value of the amount. See also the
// String() method.
func (p Percentage) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// MarshalJSON provides the text value of percentage wrapped in
// quotes ready to be included in a JSON object.
func (p Percentage) MarshalJSON() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.WriteByte('"')
	buf.WriteString(p.String())
	buf.WriteByte('"')
	return buf.Bytes(), nil
}

// UnmarshalText will decode the percentage value, even if it is quoted
// as a string.
func (p *Percentage) UnmarshalText(value []byte) error {
	if string(value) == "null" {
		return nil
	}

	str := unquote(value)
	result, err := PercentageFromString(string(str))
	if err != nil {
		return fmt.Errorf("decoding string `%s`: %w", str, err)
	}
	*p = result

	return nil
}

// UnmarshalJSON ensures we parse percentage numbers correctly from a JSON
// source.
func (p *Percentage) UnmarshalJSON(value []byte) error {
	return p.UnmarshalText(value)
}

// JSONSchema provides a representation of the struct for usage in Schema.
func (Percentage) JSONSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:        "string",
		Pattern:     `^\-?[0-9]+(\.[0-9]+)?%$`,
		Title:       "Percentage",
		Description: "Similar to an Amount, but designed for percentages and includes % symbol in JSON output.",
	}
}
