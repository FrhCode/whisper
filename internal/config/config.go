package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	FFmpeg           string      `json:"ffmpeg"`
	Microphone       string      `json:"microphone"`
	AutoPaste        bool        `json:"autoPaste"`
	ClipboardRestore bool        `json:"clipboardRestore"`
	Cloud            CloudConfig `json:"cloud"`
	LLM              LLMConfig   `json:"llm"`
}

type CloudConfig struct {
	URL      string `json:"url"`
	APIKey   string `json:"apiKey"`
	Model    string `json:"model"`
	Language string `json:"language"`
}

type LLMConfig struct {
	Enabled     bool    `json:"enabled"`
	URL         string  `json:"url"`
	APIKey      string  `json:"apiKey"`
	Model       string  `json:"model"`
	Temperature float64 `json:"temperature"`
}

func Default() Config {
	return Config{
		FFmpeg:           "bin/ffmpeg.exe",
		Microphone:       "default",
		AutoPaste:        true,
		ClipboardRestore: true,
		Cloud: CloudConfig{
			URL:      "https://router.farhandev.my.id/v1/audio/transcriptions",
			APIKey:   "",
			Model:    "dg/nova-3",
			Language: "id",
		},
		LLM: LLMConfig{
			Enabled:     true,
			URL:         "https://router.farhandev.my.id/v1/chat/completions",
			APIKey:      "",
			Model:       "",
			Temperature: 0.1,
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, Save(path, cfg)
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0644)
}
