package mx_test

import (
	"testing"

	"github.com/invopop/gobl/l10n"
	"github.com/invopop/gobl/org"
	"github.com/invopop/gobl/regimes/mx"
	"github.com/invopop/gobl/tax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigratePartyIdentities(t *testing.T) {
	customer := &org.Party{
		Name: "Test Customer",
		Identities: []*org.Identity{
			{
				Key:  mx.ExtKeyCFDIFiscalRegime,
				Code: "608",
			},
			{
				Key:  mx.ExtKeyCFDIUse,
				Code: "G01",
			},
		},
		TaxID: &tax.Identity{
			Country: l10n.MX,
			Code:    "ZZZ010101ZZZ",
			Zone:    "65000",
		},
	}

	err := customer.Calculate()
	require.NoError(t, err)

	assert.Empty(t, customer.Identities)
	assert.Len(t, customer.Ext, 2)
	assert.Equal(t, "608", customer.Ext[mx.ExtKeyCFDIFiscalRegime].String())
	assert.Equal(t, "G01", customer.Ext[mx.ExtKeyCFDIUse].String())
}
