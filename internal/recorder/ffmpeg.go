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
	if b, err := exec.Command(r.ffmpeg, "-version").CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg not runnable: %w: %s", err, b)
	}

	args := []string{"-hide_banner", "-y"}
	if r.mic == "" || r.mic == "default" {
		args = append(args, "-f", "wasapi", "-i", "default")
	} else {
		args = append(args, "-f", "dshow", "-i", "audio="+r.mic)
	}
	args = append(args, "-ac", "1", "-ar", "16000", r.out)
	r.cmd = exec.CommandContext(ctx, r.ffmpeg, args...)
	r.cmd.Stderr = &r.stderr
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.stdin = stdin
	if err := r.cmd.Start(); err != nil {
		return err
	}
	time.Sleep(300 * time.Millisecond)
	if r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		return fmt.Errorf("ffmpeg exited early: %s", r.stderr.String())
	}
	return nil
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
