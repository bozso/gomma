package date

import (
    "time"
)

type Dater interface {
    Date() time.Time
}

func Before(one, two Dater) bool {
    return one.Date().Before(two.Date())
}

type Ranger interface {
    Start() time.Time
    Center() time.Time
    Stop() time.Time
}
