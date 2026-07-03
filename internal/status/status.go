package status

import "log"

type Sink interface {
	Recording()
	Processing()
	Pasted()
	Error(string)
}

type Stdout struct{}

func (Stdout) Recording()     { log.Println("Recording") }
func (Stdout) Processing()    { log.Println("Transcribing...") }
func (Stdout) Pasted()        { log.Println("Pasted") }
func (Stdout) Error(s string) { log.Println("Error:", s) }
