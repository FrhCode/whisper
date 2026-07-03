package status

import (
	"time"

	"whispr/internal/overlay"
)

type Overlay struct{ O *overlay.Overlay }

func (s Overlay) Recording()  { s.O.Set("● Recording") }
func (s Overlay) Processing() { s.O.Set("Transcribing...") }
func (s Overlay) Pasted() {
	s.O.Set("✓ Pasted")
	go func() { time.Sleep(time.Second); s.O.Hide() }()
}
func (s Overlay) Error(msg string) { s.O.Set("Error: " + msg) }
