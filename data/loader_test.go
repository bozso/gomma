package data

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	"github.com/bozso/gomma/date"
)

var logger = log.Default()

type TestPayload struct {
	path string
	meta string
}

func (t TestPayload) MarshalJSON() (b []byte, err error) {
	return json.Marshal(map[string]string{
		"data_path": t.path,
		"meta":      t.meta,
	})
}

func TestLoadMeta(t *testing.T) {
	date_, err := date.DefaultParser.ParseDate("2018-08-09")
	if err != nil {
		t.Fatalf("date parsing failed: %s", err)
	}

	data := File{
		DataFile: New("data.mli"),
		Meta: Meta{
			DataType: KindFloat,
			RngAzi: RngAzi{
				Rng: 456,
				Azi: 128,
			},
			Date:      date.New(date_),
			CreatedBy: CreationUnknown(),
		},
	}

	marshaled := []byte(`
    {
        "path": "data.mli",
        "meta": {
            "data_type": "FLOAT",
            "range_azimuth": {
                "range": 456,
                "azimuth": 128
            },
            "date": "2018-08-09"
        }
    }`)
	logger.Printf("payload: '%s'\n", string(marshaled))

	var parsed File
	if err := json.Unmarshal(marshaled, &parsed); err != nil {
		t.Fatalf("json unmarshaling failed: %s", err)
	}

	if !reflect.DeepEqual(parsed, data) {
		t.Fatalf("expected parsed '%#v' and original '%#v' to be equal", parsed, data)
	}
}
