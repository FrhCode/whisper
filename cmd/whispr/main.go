package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/getlantern/systray"

	"whispr/internal/config"
	"whispr/internal/dictation"
	"whispr/internal/hotkey"
	"whispr/internal/icon"
	"whispr/internal/overlay"
	"whispr/internal/status"
)

func main() {
	useExeDir()
	if len(os.Args) > 1 && os.Args[1] == "dict" {
		runOnce()
		return
	}

	systray.Run(onReady, func() {})
}

func useExeDir() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	dir := filepath.Dir(exe)
	if strings.Contains(dir, "go-build") {
		return
	}
	_ = os.Chdir(dir)
}

func runOnce() {
	cfg, err := config.Load("config.json")
	must(err)
	d := dictation.New(cfg, status.Stdout{})
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	must(d.Start(ctx))
	fmt.Println("Recording. Press Enter to stop.")
	_, _ = fmt.Scanln()
	_, err = d.Stop(ctx)
	must(err)
}

func onReady() {
	systray.SetIcon(icon.Data())
	systray.SetTitle("Whispr")
	systray.SetTooltip("Whispr dictation")

	mStart := systray.AddMenuItem("Start Dictation", "Start/stop dictation")
	mMic := systray.AddMenuItem("Microphone", "Edit microphone in config.json")
	mModel := systray.AddMenuItem("Model", "Edit model in config.json")
	mHotkey := systray.AddMenuItem("Hotkey: Ctrl+Alt+Space", "Fixed for now")
	mAutoPaste := systray.AddMenuItemCheckbox("Auto Paste", "Toggle auto paste", true)
	mRestore := systray.AddMenuItemCheckbox("Clipboard Restore", "Restore clipboard after paste", true)
	mLogs := systray.AddMenuItem("Open Logs", "Open logs folder")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit")

	cfg, err := config.Load("config.json")
	if err != nil {
		log.Println(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	ov := overlay.New(ctx)
	st := status.Multi{status.Stdout{}, status.Overlay{O: ov}}
	mAutoPaste.Check()
	if !cfg.AutoPaste {
		mAutoPaste.Uncheck()
	}
	mRestore.Check()
	if !cfg.ClipboardRestore {
		mRestore.Uncheck()
	}
	d := dictation.New(cfg, st)

	go func() {
		if err := hotkey.Listen(ctx, func() { toggle(ctx, d) }); err != nil {
			st.Error(err.Error())
		}
	}()

	go func() {
		for {
			select {
			case <-mStart.ClickedCh:
				toggle(ctx, d)
			case <-mMic.ClickedCh:
				_ = openPath("config.json")
			case <-mModel.ClickedCh:
				_ = openPath("config.json")
			case <-mHotkey.ClickedCh:
				st.Error("Hotkey fixed: Ctrl+Alt+Space")
			case <-mAutoPaste.ClickedCh:
				cfg.AutoPaste = !cfg.AutoPaste
				setCheck(mAutoPaste, cfg.AutoPaste)
				_ = config.Save("config.json", cfg)
				d.SetConfig(cfg)
			case <-mRestore.ClickedCh:
				cfg.ClipboardRestore = !cfg.ClipboardRestore
				setCheck(mRestore, cfg.ClipboardRestore)
				_ = config.Save("config.json", cfg)
				d.SetConfig(cfg)
			case <-mLogs.ClickedCh:
				_ = os.MkdirAll("logs", 0755)
				_ = openPath(filepath.Join(".", "logs"))
			case <-mQuit.ClickedCh:
				cancel()
				systray.Quit()
				return
			}
		}
	}()
}

type checkItem interface {
	Check()
	Uncheck()
}

func setCheck(m checkItem, checked bool) {
	if checked {
		m.Check()
		return
	}
	m.Uncheck()
}

func toggle(ctx context.Context, d *dictation.Dictation) {
	if d.Recording() {
		_, _ = d.Stop(ctx)
		return
	}
	if err := d.Start(ctx); err != nil {
		log.Println(err)
	}
}

func openPath(path string) error {
	return execCommand("explorer", path)
}

func execCommand(name string, args ...string) error {
	p, err := os.StartProcess(name, append([]string{name}, args...), &os.ProcAttr{Files: []*os.File{nil, os.Stdout, os.Stderr}})
	if err != nil {
		return err
	}
	_, err = p.Wait()
	return err
}

func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var _ = time.Second
