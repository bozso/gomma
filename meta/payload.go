package meta

import (
    "fmt"
    "encoding/json"
)

type Payload map[string]json.RawMessage

func (p Payload) HasTag(tag string) (b bool) {
    _, b = p[tag]
    return
}

func (p Payload) MustHaveTag(tag string) (msg json.RawMessage, err error) {
    err = nil
    msg, ok := p[tag]
    if !ok {
        err = MissingTag{tag}
    }
    return
}

type MissingTag struct {
    Tag string
}

func (p Payload) Unmarshal(tag string, v interface{}) (err error) {
    msg, err := p.MustHaveTag(tag)
    if err != nil {
        return
    }

    return json.Unmarshal(msg, v)
}

func (m MissingTag) Error() (s string) {
    return fmt.Sprintf("tag '%s' was not found for payload", m.Tag)
}
