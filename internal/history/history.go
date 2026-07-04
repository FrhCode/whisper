package history

import (
	"encoding/json"
	"os"
	"time"
)

type Entry struct {
	Time    string `json:"time"`
	Raw     string `json:"raw"`
	Cleaned string `json:"cleaned"`
}

func Append(path, raw, cleaned string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.Marshal(Entry{Time: time.Now().Format(time.RFC3339), Raw: raw, Cleaned: cleaned})
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	return err
}
