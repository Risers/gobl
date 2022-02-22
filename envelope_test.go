package gobl_test

import (
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/invopop/gobl"
	"github.com/invopop/gobl/bill"
	"github.com/invopop/gobl/dsig"
	"github.com/invopop/gobl/note"
	"github.com/invopop/gobl/uuid"
)

var testKey = dsig.NewES256Key()

func TestEnvelopeDocument(t *testing.T) {
	m := new(note.Message)
	m.Content = "This is test content."

	e := gobl.NewEnvelope()
	if assert.NotNil(t, e.Head) {
		assert.NotEmpty(t, e.Head.UUID, "empty header uuid")
	}
	assert.NotNil(t, e.Document)

	if err := e.Insert(m); err != nil {
		t.Errorf("failed to insert payload: %v", err)
		return
	}

	if assert.NotNil(t, e.Head.Digest) {
		assert.Equal(t, e.Head.Digest.Algorithm, dsig.DigestSHA256, "unexpected digest algorithm")
		assert.Equal(t, "c6a5148ce90f70c24ebfe6de1abed0d0aafde4323a9bcf47cc4a5d544af9ea19", e.Head.Digest.Value, "digest should be the same")
	}

	assert.Empty(t, e.Signatures)
	assert.NoError(t, e.Sign(testKey), "signing envelope")
	assert.NotEmpty(t, e.Signatures, "expected a signature")

	assert.NoError(t, e.Verify(), "did not expect verify error")

	nm, ok := e.Extract().(*note.Message)
	require.True(t, ok, "unrecognized content")
	assert.Equal(t, m.Content, nm.Content, "content mismatch")
}

func TestEnvelopeExtract(t *testing.T) {
	e := &gobl.Envelope{}
	obj := e.Extract()
	assert.Nil(t, obj)
}

func TestEnvelopeComplete(t *testing.T) {
	e := new(gobl.Envelope)

	data, err := ioutil.ReadFile("./samples/envelope-invoice-es.yaml")
	require.NoError(t, err)
	err = yaml.Unmarshal(data, e)
	require.NoError(t, err)

	err = e.Complete()
	require.NoError(t, err)

	inv, ok := e.Extract().(*bill.Invoice)
	require.True(t, ok)
	require.NoError(t, err)

	assert.Equal(t, "1210.00", inv.Totals.Payable.String())
}

func TestEnvelopeValidate(t *testing.T) {
	tests := []struct {
		name string
		env  *gobl.Envelope
		want string
	}{
		{
			name: "no head nor version",
			env:  &gobl.Envelope{},
			want: "$schema: cannot be blank; doc: cannot be blank; head: cannot be blank.",
		},
		{
			name: "missing sig, draft",
			env: &gobl.Envelope{
				Schema: gobl.EnvelopeSchema,
				Head: &gobl.Header{
					Digest: &dsig.Digest{},
					Draft:  true,
					UUID:   uuid.NewV1(),
				},
				Document: new(gobl.Document),
			},
		},
		{
			name: "missing sig, draft",
			env: &gobl.Envelope{
				Schema: gobl.EnvelopeSchema,
				Head: &gobl.Header{
					Digest: &dsig.Digest{},
					UUID:   uuid.NewV1(),
				},
				Document: new(gobl.Document),
			},
			want: "sigs: cannot be blank.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.env.Validate()
			if tt.want == "" && err == nil {
				return
			}
			assert.EqualError(t, err, tt.want)
		})
	}
}
