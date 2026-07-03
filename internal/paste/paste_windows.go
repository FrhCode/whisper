//go:build windows

package paste

import (
	"context"
	"os/exec"
	"strings"
)

func Text(ctx context.Context, s string, restore bool) error {
	ps := `
$old = Get-Clipboard -Raw -ErrorAction SilentlyContinue
$t = @'
` + strings.ReplaceAll(s, "'@", "' + \"@\" + '") + `
'@
Set-Clipboard -Value $t
$ws = New-Object -ComObject WScript.Shell
$ws.SendKeys('^v')
Start-Sleep -Milliseconds 350
if (` + boolPS(restore) + `) { Set-Clipboard -Value $old }
`
	return exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", ps).Run()
}

func boolPS(v bool) string {
	if v {
		return "$true"
	}
	return "$false"
}
