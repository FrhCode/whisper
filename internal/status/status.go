package status

import "log"

type Sink interface {
	Recording()
	Processing()
	Cleaning()
	Pasted()
	Error(string)
}

type Stdout struct{}

func (Stdout) Recording()     { log.Println("Recording") }
func (Stdout) Processing()    { log.Println("Transcribing...") }
func (Stdout) Cleaning()      { log.Println("Cleaning text...") }
func (Stdout) Pasted()        { log.Println("Pasted") }
func (Stdout) Error(s string) { log.Println("Error:", s) }
