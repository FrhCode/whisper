# Whispr

Windows-only tray dictation prototype.

## Layout

Place bundled file:

```text
bin/ffmpeg.exe
```

First run creates `config.json`.

## Run

```powershell
go mod tidy
go run ./cmd/whispr
```

First run creates `config.json`. Put API key in:

```json
"cloud": {
  "url": "https://router.farhandev.my.id/v1/audio/transcriptions",
  "apiKey": "xxxxxxxxxxxxxxxxxxxxxxx",
  "model": "dg/nova-3",
  "language": "id"
},
"llm": {
  "enabled": true,
  "url": "https://router.farhandev.my.id/v1/chat/completions",
  "apiKey": "xxxxxxxxxxxxxxxxxxxxxxx",
  "model": "YOUR_LLM_MODEL",
  "temperature": 0.1
}
```

Tray starts. Press `Ctrl+Alt+Space` to start recording, press again to stop, upload, transcribe, paste.

CLI fallback:

```powershell
go run ./cmd/whispr dict
```

## Microphone

`config.json` default uses:

```json
"microphone": "default"
```

If ffmpeg cannot open default mic, list devices:

```powershell
bin/ffmpeg.exe -list_devices true -f dshow -i dummy
```

Copy exact audio device name into `config.json`, for example:

```json
"microphone": "Microphone Array"
```
