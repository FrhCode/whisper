package recorder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

type Recorder struct {
	ffmpeg string
	mic    string
	out    string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stderr bytes.Buffer
}

func New(ffmpeg, mic, out string) *Recorder {
	return &Recorder{ffmpeg: ffmpeg, mic: mic, out: out}
}

func (r *Recorder) Start(ctx context.Context) error {
	input := "audio=" + r.mic
	if r.mic == "" || r.mic == "default" {
		input = "audio=default"
	}
	if b, err := exec.Command(r.ffmpeg, "-version").CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg not runnable: %w: %s", err, b)
	}

	r.cmd = exec.CommandContext(ctx, r.ffmpeg, "-hide_banner", "-y", "-f", "dshow", "-i", input, "-ac", "1", "-ar", "16000", r.out)
	r.cmd.Stderr = &r.stderr
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.stdin = stdin
	return r.cmd.Start()
}

func (r *Recorder) Stop() error {
	if r.cmd == nil {
		return nil
	}
	_, _ = io.WriteString(r.stdin, "q\n")
	_ = r.stdin.Close()

	done := make(chan error, 1)
	go func() { done <- r.cmd.Wait() }()

	var err error
	select {
	case err = <-done:
	case <-time.After(2 * time.Second):
		_ = r.cmd.Process.Kill()
		err = <-done
	}

	if okWAV(r.out) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("ffmpeg stop failed: %w: %s", err, r.stderr.String())
	}
	return fmt.Errorf("ffmpeg produced no audio: %s", r.stderr.String())
}

func okWAV(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 44
}
