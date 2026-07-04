package dictation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"whispr/internal/beep"
	"whispr/internal/cleanup"
	"whispr/internal/config"
	"whispr/internal/paste"
	"whispr/internal/recorder"
	"whispr/internal/status"
	"whispr/internal/transcribe"
)

type Dictation struct {
	cfg config.Config
	st  status.Sink
	mu  sync.Mutex
	rec *recorder.Recorder
}

func New(cfg config.Config, st status.Sink) *Dictation { return &Dictation{cfg: cfg, st: st} }

func (d *Dictation) SetConfig(cfg config.Config) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.cfg = cfg
}

func (d *Dictation) Recording() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rec != nil
}

func (d *Dictation) Start(ctx context.Context) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.rec != nil {
		return nil
	}
	if err := os.MkdirAll("tmp", 0755); err != nil {
		return err
	}
	r := recorder.New(d.cfg.FFmpeg, d.cfg.Microphone, filepath.Join("tmp", "temp.wav"))
	if err := r.Start(ctx); err != nil {
		d.st.Error(err.Error())
		beep.Error()
		return err
	}
	d.rec = r
	beep.Start()
	d.st.Recording()
	return nil
}

func (d *Dictation) Stop(ctx context.Context) (string, error) {
	d.mu.Lock()
	r := d.rec
	d.rec = nil
	d.mu.Unlock()
	if r == nil {
		return "", nil
	}
	beep.Stop()
	if err := r.Stop(); err != nil {
		d.st.Error(err.Error())
		beep.Error()
		return "", err
	}
	d.st.Processing()
	text, err := transcribe.Run(ctx, d.cfg.Cloud, filepath.Join("tmp", "temp.wav"))
	if err != nil {
		d.st.Error(err.Error())
		beep.Error()
		return "", err
	}
	if text == "" {
		return "", fmt.Errorf("empty transcript")
	}
	cleaned, err := cleanup.Run(ctx, d.cfg.LLM, text)
	if err != nil {
		d.st.Error(err.Error())
		beep.Error()
	} else {
		text = cleaned
	}
	if text == "" {
		return "", nil
	}
	if d.cfg.AutoPaste {
		if err := paste.Text(ctx, text, d.cfg.ClipboardRestore); err != nil {
			d.st.Error("paste failed; transcript copied nowhere: " + err.Error())
			beep.Error()
			return text, err
		}
		d.st.Pasted()
	}
	return text, nil
}
