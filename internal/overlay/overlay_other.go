//go:build !windows

package overlay

import "context"

type Overlay struct{}

func New(context.Context) *Overlay { return &Overlay{} }
func (*Overlay) Set(string)        {}
func (*Overlay) Hide()             {}
