package meta

import (
    "testing"
    "strings"
    "encoding/json"
)

type dataA struct {
    Str string `json:"str"`
    Int int    `json:"int"`
}

type dataB struct {
    DataA dataA `json:"data_a"`
    Float float64 `json:"float"`
}

func TestPayload(t *testing.T) {
    const payload = `
    {
        "a": {
            "str": "b",
            "int": 5
        }
    }
    `

    var pl Payload
    err := json.NewDecoder(strings.NewReader(payload)).Decode(&pl)
    if err != nil {
        t.Fatalf("unmarshal failed: %s", err)
    }

    data, expected := dataA{}, dataA{
        Str: "b", Int: 5,
    }

    if err = pl.Unmarshal("a", &data); err != nil {
        t.Fatalf("unmarshal failed: %s", err)
    }

    if data != expected {
        t.Errorf("%#v and %#v should be equal", data, expected)
    }
}

func TestNested(t *testing.T) {
    const payload = `
    {
        "data": {
            "b": {
                "float": 6.5,
                "data_a": {
                    "str": "aaa",
                    "int": 6
                }
            }
        }
    }
    `
    var pl = struct {
        Data Payload `json:"data"`
    }{}

    err := json.NewDecoder(strings.NewReader(payload)).Decode(&pl)
    if err != nil {
        t.Errorf("unmarshal failed: %s", err)
    }

    data, expected := dataB{}, dataB{
        DataA: dataA{
            Str: "aaa",
            Int: 6,
        },
        Float: 6.5,
    }

    if err = pl.Data.Unmarshal("b", &data); err != nil {
        t.Fatalf("unmarshal failed: %s", err)
    }

    if data != expected {
        t.Errorf("%#v and %#v should be equal", data, expected)
    }
}
