package tax

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/invopop/gobl/cbc"
	"github.com/invopop/gobl/i18n"
	"github.com/invopop/gobl/l10n"
	"github.com/invopop/validation"
)

// Zone represents an area inside a country, like a province
// or a state, which shares the basic definitions of the country, but
// may vary in some validation rules.
type Zone struct {
	// Unique zone code.
	Code l10n.Code `json:"code" jsonschema:"title=Code"`
	// Name of the zone to be use if a locality or region is not applicable.
	Name i18n.String `json:"name,omitempty" jsonschema:"title=Name"`
	// Village, town, district, or city name which should coincide with
	// address data.
	Locality i18n.String `json:"locality,omitempty" jsonschema:"title=Locality"`
	// Province, county, or state which should match address data.
	Region i18n.String `json:"region,omitempty" jsonschema:"title=Region"`
	// Codes defines a set of regime specific code mappings.
	Codes cbc.CodeMap `json:"codes,omitempty" jsonschema:"title=Codes"`
	// Any additional information
	Meta cbc.Meta `json:"meta,omitempty" jsonschema:"title=Meta"`
}

// Validate ensures that the zone looks correct.
func (z *Zone) Validate() error {
	err := validation.ValidateStruct(z,
		validation.Field(&z.Code, validation.Required),
		validation.Field(&z.Name),
		validation.Field(&z.Locality),
		validation.Field(&z.Region),
		validation.Field(&z.Meta),
	)
	return err
}

type validateZoneCode struct {
	store *ZoneStore
}

// Validate checks to see if the provided zone appears in the store.
func (v *validateZoneCode) Validate(value interface{}) error {
	code, ok := value.(l10n.Code)
	if !ok || code == "" {
		return nil
	}
	if z := v.store.Get(code); z == nil {
		return errors.New("must be a valid value")
	}
	return nil
}

// ZoneIn returns a validation rule that checks to see if the provided
// zone is in the store.
func ZoneIn(store *ZoneStore) validation.Rule {
	return &validateZoneCode{store}
}

// ZoneStore makes it easier to load zone information dynamically from
// source data.
type ZoneStore struct {
	sync.Mutex
	src  embed.FS
	fn   string
	data struct {
		Zones []*Zone `json:"zones"`
	}
}

// NewZoneStore instantiates a new zone store that will use and embedded
// file system for loading the data.
func NewZoneStore(fs embed.FS, filename string) *ZoneStore {
	return &ZoneStore{src: fs, fn: filename}
}

// JSONSchemaAlias provides the real object that should be defined in the schemas.
func (ZoneStore) JSONSchemaAlias() any { //nolint:govet
	return []*Zone{}
}

func (s *ZoneStore) load() {
	s.Lock()
	defer s.Unlock()

	if len(s.data.Zones) == 0 {
		data, err := s.src.ReadFile(s.fn)
		if err != nil {
			panic(fmt.Sprintf("expected to find zone data: %s", err))
		}
		s.data.Zones = make([]*Zone, 0)
		if err := json.Unmarshal(data, &s.data); err != nil {
			panic(fmt.Sprintf("parsing zone data: %s", err))
		}
	}
}

// Get will load the zone object from the JSON data.
func (s *ZoneStore) Get(code l10n.Code) *Zone {
	s.load()
	for _, z := range s.data.Zones {
		if z.Code == code {
			return z
		}
	}
	return nil
}

// Codes provides the list of available zone codes.
func (s *ZoneStore) Codes() []l10n.Code {
	s.load()
	codes := make([]l10n.Code, len(s.data.Zones))
	for i, z := range s.data.Zones {
		codes[i] = z.Code
	}
	return codes
}

// List provides the complete zone list.
func (s *ZoneStore) List() []*Zone {
	return s.data.Zones
}

// MarshalJSON ensures the zone data is loaded before marshaling.
func (s *ZoneStore) MarshalJSON() ([]byte, error) {
	s.load()
	return json.Marshal(s.data.Zones)
}
