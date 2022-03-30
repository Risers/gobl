package tax

import (
	"errors"
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-playground/validator/v10"

	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/num"
	"github.com/invopop/gobl/org"
)

// Region defines the holding structure for a regions categories and subsequent
// Rates and Values.
type Region struct {
	Code       Code        `json:"code" jsonschema:"title=Code" validate:"required"`
	Name       i18n.String `json:"name" jsonschema:"title=Name" validate:"required"`
	Categories []Category  `json:"categories" jsonschema:"title=Categories" validate:"required"`
}

// Category ...
type Category struct {
	Code Code        `json:"code" jsonschema:"title=Code"`
	Name i18n.String `json:"name" jsonschema:"title=Name"`
	Desc i18n.String `json:"desc,omitempty" jsonschema:"title=Description"`

	// Retained when true implies that the tax amount will be retained
	// by the buyer on behalf of the supplier, and thus subtracted from
	// the invoice taxable base total. Typically used for taxes related to
	// income.
	Retained bool `json:"retained,omitempty" jsonschema:"title=Retained"`

	// Specific tax definitions inside this category.
	Defs []Def `json:"defs" jsonschema:"title=Definitions"`
}

// Def defines a tax combination of category and rate.
type Def struct {
	// Code identifies this rate within the system
	Code Code `json:"code" jsonschema:"title=Code"`

	Name i18n.String `json:"name" jsonschema:"title=Name"`
	Desc i18n.String `json:"desc,omitempty" jsonschema:"title=Description"`

	// Values contains a list of Value objects that contain the
	// current and historical percentage values for the rate.
	// Order is important, newer values should come before
	// older values.
	Values []Value `json:"values" jsonschema:"title=Values,description=Set of values ordered by date that determine what rates to apply since when."`
}

// Value contains a percentage rate or fixed amount for a given date range.
// Fiscal policy changes mean that rates are not static so we need to
// be able to apply the correct rate for a given period.
type Value struct {
	// Date from which this value should be applied.
	Since *org.Date `json:"since,omitempty" jsonschema:"title=Since"`
	// Rate that should be applied
	Percent num.Percentage `json:"percent" jsonschema:"title=Percent"`
	// When true, this value should no longer be used.
	Disabled bool `json:"disabled,omitempty" jsonschema:"title=Disabled"`
}

// combo is used internally to make it easier to return a final value including
// all the preciding objects.
type combo struct {
	category Category
	def      Def
	value    Value
}

// Validate enures the region definition is valid, including all
// subsequent categories.
func (r Region) Validate() error {
	return validator.New().Struct(r)
}

// Validate ensures the Category's contents are correct.
func (c Category) Validate() error {
	err := validation.ValidateStruct(&c,
		validation.Field(&c.Code, validation.Required),
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Defs, validation.Required),
	)
	return err
}

// Validate checks that our tax definition is valid. This is only really
// meant to be used when testing new regional tax definitions.
func (d Def) Validate() error {
	err := validation.ValidateStruct(&d,
		validation.Field(&d.Code, validation.Required),
		validation.Field(&d.Name, validation.Required),
		validation.Field(&d.Values, validation.Required, validation.By(checkDefValuesOrder)),
	)
	return err
}

// Validate ensures the tax rate contains all the required fields.
func (v Value) Validate() error {
	return validation.ValidateStruct(&v,
		validation.Field(&v.Percent, validation.Required),
	)
}

func checkDefValuesOrder(list interface{}) error {
	values, ok := list.([]Value)
	if !ok {
		return errors.New("must be a tax rate value array")
	}
	var date *org.Date
	// loop through and check order of Since value
	for i := range values {
		v := &values[i]
		if date != nil && date.IsValid() {
			if v.Since.IsValid() && !v.Since.Before(date.Date) {
				return errors.New("invalid date order")
			}
		}
		date = v.Since
	}
	return nil
}

// Category provides the requested category by its code.
func (r Region) Category(code Code) (Category, bool) {
	for _, c := range r.Categories {
		if c.Code == code {
			return c, true
		}
	}
	return Category{}, false
}

// Def provides the rate definition with a matching code for
// the category.
func (c Category) Def(code Code) (Def, bool) {
	for _, d := range c.Defs {
		if d.Code == code {
			return d, true
		}
	}
	return Def{}, false
}

// On determines the tax rate value for the provided date.
func (d Def) On(date org.Date) (Value, bool) {
	for _, v := range d.Values {
		if !v.Since.IsValid() || v.Since.Before(date.Date) {
			return v, true
		}
	}
	return Value{}, false
}

// comboOn provides the Value object for the provided rate on a given day
// or an error if no match is found.
func (r Region) comboOn(rate *Rate, date org.Date) (*combo, error) {
	c := new(combo)
	var ok bool
	c.category, ok = r.Category(rate.Category)
	if !ok {
		return nil, fmt.Errorf("failed to find category, invalid code: %v", rate.Category)
	}
	c.def, ok = c.category.Def(rate.Code)
	if !ok {
		return nil, fmt.Errorf("failed to find rate definition, invalid code: %v", rate.Code)
	}
	c.value, ok = c.def.On(date)
	if !ok {
		return nil, fmt.Errorf("tax rate cannot be provided for date")
	}
	return c, nil
}
