package recorder

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

type Recorder struct {
	ffmpeg string
	mic    string
	out    string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
}

func New(ffmpeg, mic, out string) *Recorder {
	return &Recorder{ffmpeg: ffmpeg, mic: mic, out: out}
}

func (r *Recorder) Start(ctx context.Context) error {
	input := "audio=" + r.mic
	if r.mic == "" || r.mic == "default" {
		input = "audio=default"
	}
	r.cmd = exec.CommandContext(ctx, r.ffmpeg, "-y", "-f", "dshow", "-i", input, "-ac", "1", "-ar", "16000", r.out)
	stdin, err := r.cmd.StdinPipe()
	if err != nil {
		return err
	}
	r.stdin = stdin
	if b, err := exec.Command(r.ffmpeg, "-version").CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg not runnable: %w: %s", err, b)
	}
	return r.cmd.Start()
}

func (r *Recorder) Stop() error {
	if r.cmd == nil {
		return nil
	}
	_, _ = io.WriteString(r.stdin, "q")
	return r.cmd.Wait()
}
