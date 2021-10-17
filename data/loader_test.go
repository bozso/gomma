package data

import (
	"testing"
)

type TestPayload struct {
	path string
	meta string
}

var testPayload = TestPayload{
	path: "data.mli",
	meta: `
{
    "data_type": "FLOAT",
    "range_azimuth": {
        "range": 456,
        "azimuth": 128,
    },
    "date": 2018.08.09
}`,
}

func TestLoadMeta(t *testing.T) {
}
