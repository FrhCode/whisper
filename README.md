# Whispr

Windows tray dictation app.

Press hotkey, speak, press hotkey again. Whispr transcribes Indonesian speech, cleans text with LLM, then pastes into current app.

## Requirements

- Windows
- Go
- ffmpeg installed globally
- Router API key

## Install requirements

Open PowerShell:

```powershell
winget install GoLang.Go
winget install Gyan.FFmpeg
```

Close PowerShell, open again, check:

```powershell
go version
ffmpeg -version
```

## Build

From project root:

```powershell
.\scripts\build-windows.ps1
```

Output:

```text
dist\whispr.exe
dist\config.example.json
```

## Configure

```powershell
cd dist
copy config.example.json config.json
notepad config.json
```

Fill:

```text
cloud.apiKey
llm.apiKey
llm.model
```

Example:

```json
{
  "ffmpeg": "ffmpeg",
  "microphone": "default",
  "autoPaste": true,
  "clipboardRestore": true,
  "cloud": {
    "url": "https://router.farhandev.my.id/v1/audio/transcriptions",
    "apiKey": "YOUR_API_KEY",
    "model": "dg/nova-3",
    "language": "id"
  },
  "llm": {
    "enabled": true,
    "url": "https://router.farhandev.my.id/v1/chat/completions",
    "apiKey": "YOUR_API_KEY",
    "model": "YOUR_LLM_MODEL",
    "temperature": 0.1
  }
}
```

## Run

```powershell
.\whispr.exe
```

Or double-click:

```text
dist\whispr.exe
```

Keep these together if moving app:

```text
Whispr/
  whispr.exe
  config.json
```

Whispr reads `config.json` next to `whispr.exe`.

## Use

```text
Ctrl+Alt+Space  start recording
Ctrl+Alt+Space  stop recording
```

Flow:

```text
Recording
Transcribing...
Cleaning text...
Pasted
```

## Microphone

Default config:

```json
"microphone": "default"
```

Whispr auto-detects first DirectShow audio device.

If wrong mic, list devices:

```powershell
ffmpeg -hide_banner -list_devices true -f dshow -i dummy
```

Copy exact audio device name:

```json
"microphone": "Microphone Array (AMD Audio Device)"
```

## Quit

Right-click tray icon:

```text
Quit
```
