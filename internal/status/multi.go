package status

type Multi []Sink

func (m Multi) Recording() {
	for _, s := range m {
		s.Recording()
	}
}
func (m Multi) Processing() {
	for _, s := range m {
		s.Processing()
	}
}
func (m Multi) Pasted() {
	for _, s := range m {
		s.Pasted()
	}
}
func (m Multi) Error(msg string) {
	for _, s := range m {
		s.Error(msg)
	}
}
