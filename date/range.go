package date

import (
    "time"
)

type Range struct {
    start, stop, center time.Time
}

func NewRange(start, stop time.Time) (r Range) {
    // TODO: Optional check duration, is it max or min
    delta := start.Sub(stop) / 2.0
    r.center = stop.Add(delta)

    r.start = start
    r.stop = stop    
}

func (df ParseFmt) NewRange(start, stop string) (d Range, err error) {
    dstart, err := df.Parse(start)
    if err != nil {
        return
    }

    dstop, err := df.Parse(stop)
    if err != nil {
        return
    }

    d = NewRange(dstart, dstop)
    
    return d, nil
}

func (d Range) Start() time.Time {
    return d.start
}

func (d Range) Center() time.Time {
    return d.center
}

func (d Range) Stop() time.Time {
    return d.stop
}
