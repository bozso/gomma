package date

import (
    "time"
)

type Checker interface {
    In(time.Time) bool
}

type Range struct {
    start, stop, center time.Time
}

func NewRange(start, stop time.Time) (r Range) {
    // TODO: Optional check duration, is it max or min
    delta := start.Sub(stop) / 2.0
    r.center = stop.Add(delta)

    r.start = start
    r.stop = stop
    return
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

func (d Range) In(t time.Time) (b bool) {
    return d.start.After(t) && d.stop.Before(t)
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

type MinOrMax int

const (
    Min MinOrMax = iota
    Max
)

func (m MinOrMax) New(t time.Time) (l Limit) {
    l.minOrMax, l.Time = m, t
    return
}

type Limit struct {
    time.Time
    minOrMax MinOrMax
}

func (l Limit) In(t time.Time) (b bool) {
    switch l.minOrMax {
    case Min:
        b = l.Time.After(t)
    case Max:
        b = l.Time.Before(t)
    }
    return
}

type Checkers struct {
    checkers []Checker
}

func NewCheckersWithCap(cap int) (c Checkers) {
    c.checkers = make([]Checker, 0, cap)
    return
}

func NewCheckers() (c Checkers) {
    return NewCheckersWithCap(0)
}

func (c *Checkers) Append(ch Checker) {
    c.checkers = append(c.checkers, ch)
}

func (c Checkers) In(t time.Time) (b bool) {
    if len(c.checkers) == 0 {
        return true
    }
    
    for ii, _ := range c.checkers {
        b = c.checkers[ii].In(t)
        if b {
            break
        }
    }
    return
}

var noCheck NoChecker

func NoCheck() (n NoChecker) {
    return noCheck
}

type NoChecker struct{}

func (_ NoChecker) In(_ time.Time) (b bool) {
    return true
}

func (_ NoChecker) Add(ch Checker) (c Checkers) {
    c = NewCheckersWithCap(1)
    c.Append(ch)
    return
}

func (_ NoChecker) Extend(ch ...Checker) (c Checkers) {
    c = NewCheckersWithCap(len(ch))
    for ii, _ := range ch {
        c.Append(ch[ii])
    }
    return
}
