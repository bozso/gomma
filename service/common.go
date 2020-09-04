package service

import (
    "github.com/bozso/gotoolbox/cli/stream"
)

type Empty struct{}

type Output struct {
    Out stream.Out `json:"out"`
}

