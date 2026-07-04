package recorder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
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

	mic := r.mic
	if mic == "" || mic == "default" {
		found, err := DefaultMic(ctx, r.ffmpeg)
		if err != nil {
			return err
		}
		mic = found
	}

	r.cmd = exec.CommandContext(ctx, r.ffmpeg, "-hide_banner", "-y", "-f", "dshow", "-i", "audio="+mic, "-ac", "1", "-ar", "16000", r.out)
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

func DefaultMic(ctx context.Context, ffmpeg string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	b, _ := exec.CommandContext(ctx, ffmpeg, "-hide_banner", "-list_devices", "true", "-f", "dshow", "-i", "dummy").CombinedOutput()
	devices := AudioDevices(string(b))
	if len(devices) == 0 {
		return "", fmt.Errorf("no dshow audio devices found; run: ffmpeg -list_devices true -f dshow -i dummy")
	}
	return devices[0], nil
}

func AudioDevices(s string) []string {
	var out []string
	inAudio := false
	re := regexp.MustCompile(`"([^"]+)"`)
	for _, line := range strings.Split(s, "\n") {
		if strings.Contains(line, "DirectShow audio devices") {
			inAudio = true
			continue
		}
		if strings.Contains(line, "DirectShow video devices") {
			inAudio = false
			continue
		}
		if strings.Contains(line, "Alternative name") {
			continue
		}
		m := re.FindStringSubmatch(line)
		if len(m) != 2 {
			continue
		}
		if strings.Contains(line, "(audio)") || inAudio {
			out = append(out, m[1])
		}
	}
	return out
}

func okWAV(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.Size() > 44
}
