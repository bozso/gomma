package meta

import (
    "fmt"
)

type TagNotFound struct {
    Tag string
    Choices []string
}

func (t TagNotFound) Error() (s string) {
    return fmt.Sprintf("type tag '%s' is not valid, valid options: %s",
        t.Tag, t.Choices)
}
