package cbc_test

import (
	"testing"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/validation"
	"github.com/stretchr/testify/assert"
)

func TestDuplicateStamps(t *testing.T) {
	st := struct {
		Stamps []*cbc.Stamp
	}{
		Stamps: []*cbc.Stamp{
			{
				Provider: cbc.Key("provider"),
				Value:    "value",
			},
			{
				Provider: cbc.Key("provider2"),
				Value:    "value2",
			},
		},
	}

	err := validation.ValidateStruct(&st,
		validation.Field(&st.Stamps, cbc.DetectDuplicateStamps),
	)
	assert.NoError(t, err)

	st.Stamps = append(st.Stamps, &cbc.Stamp{
		Provider: cbc.Key("provider"),
		Value:    "value3",
	})
	err = validation.ValidateStruct(&st,
		validation.Field(&st.Stamps, cbc.DetectDuplicateStamps),
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate stamp 'provider'")
}

func TestAddStamp(t *testing.T) {
	st := struct {
		Stamps []*cbc.Stamp
	}{
		Stamps: []*cbc.Stamp{
			{
				Provider: cbc.Key("provider"),
				Value:    "value",
			},
		},
	}
	st.Stamps = cbc.AddStamp(st.Stamps, &cbc.Stamp{
		Provider: cbc.Key("provider"),
		Value:    "new value",
	})
	assert.Len(t, st.Stamps, 1)
	assert.Equal(t, "new value", st.Stamps[0].Value)
}
