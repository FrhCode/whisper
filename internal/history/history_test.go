package history

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadLatest(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 55; i++ {
		_, _ = fmt.Fprintf(f, `{"time":"t%02d","raw":"r%02d","cleaned":"c%02d"}`+"\n", i, i, i)
	}
	_, _ = f.WriteString("not json\n")
	_, _ = f.WriteString(`{"time":"t56","raw":"r56","cleaned":"c56"}` + "\n")
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := LoadLatest(path, 50)
	if err != nil {
		t.Fatal(err)
	}
	if got.Skipped != 1 {
		t.Fatalf("skipped = %d, want 1", got.Skipped)
	}
	if len(got.Entries) != 50 {
		t.Fatalf("len = %d, want 50", len(got.Entries))
	}
	if got.Entries[0].Cleaned != "c56" {
		t.Fatalf("newest = %q, want c56", got.Entries[0].Cleaned)
	}
	if got.Entries[49].Cleaned != "c07" {
		t.Fatalf("oldest = %q, want c07", got.Entries[49].Cleaned)
	}
}

func TestLoadLatestMissingAndEmpty(t *testing.T) {
	got, err := LoadLatest(filepath.Join(t.TempDir(), "missing.jsonl"), 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 0 || got.Skipped != 0 {
		t.Fatalf("missing = %+v, want empty", got)
	}

	path := filepath.Join(t.TempDir(), "empty.jsonl")
	if err := os.WriteFile(path, nil, 0644); err != nil {
		t.Fatal(err)
	}
	got, err = LoadLatest(path, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Entries) != 0 || got.Skipped != 0 {
		t.Fatalf("empty = %+v, want empty", got)
	}
}
