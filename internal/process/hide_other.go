//go:build !windows

package process

import "os/exec"

func HideWindow(*exec.Cmd) {}
