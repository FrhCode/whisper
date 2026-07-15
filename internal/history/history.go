package history

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"time"
)

const DefaultLimit = 50

type Entry struct {
	Time    string `json:"time"`
	Raw     string `json:"raw"`
	Cleaned string `json:"cleaned"`
}

type LoadResult struct {
	Entries []Entry
	Skipped int
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

func LoadLatest(path string, limit int) (LoadResult, error) {
	if limit <= 0 {
		return LoadResult{}, nil
	}
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return LoadResult{}, nil
	}
	if err != nil {
		return LoadResult{}, err
	}
	defer f.Close()

	var out LoadResult
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadString('\n')
		if line != "" {
			keepLatest(&out, strings.TrimSpace(line), limit)
		}
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return LoadResult{}, err
		}
	}
	for i, j := 0, len(out.Entries)-1; i < j; i, j = i+1, j-1 {
		out.Entries[i], out.Entries[j] = out.Entries[j], out.Entries[i]
	}
	return out, nil
}

func keepLatest(out *LoadResult, line string, limit int) {
	if line == "" {
		return
	}
	var e Entry
	if err := json.Unmarshal([]byte(line), &e); err != nil {
		out.Skipped++
		return
	}
	if len(out.Entries) < limit {
		out.Entries = append(out.Entries, e)
		return
	}
	copy(out.Entries, out.Entries[1:])
	out.Entries[len(out.Entries)-1] = e
}
