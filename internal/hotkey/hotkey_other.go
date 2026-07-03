//go:build !windows

package hotkey

import (
	"context"
	"fmt"
)

func Listen(context.Context, func()) error { return fmt.Errorf("hotkey only supported on Windows") }
