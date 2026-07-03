//go:build !windows

package paste

import (
	"context"
	"fmt"
)

func Text(context.Context, string, bool) error { return fmt.Errorf("paste only supported on Windows") }
