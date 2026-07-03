package transcribe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"whispr/internal/config"
)

type response struct {
	Text string `json:"text"`
}

func Run(ctx context.Context, cfg config.CloudConfig, audio string) (string, error) {
	if cfg.URL == "" || cfg.APIKey == "" || cfg.Model == "" {
		return "", fmt.Errorf("cloud url/apiKey/model required in config.json")
	}

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	file, err := os.Open(audio)
	if err != nil {
		return "", err
	}
	defer file.Close()

	part, err := mw.CreateFormFile("file", filepath.Base(audio))
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", err
	}
	_ = mw.WriteField("model", cfg.Model)
	if cfg.Language != "" {
		_ = mw.WriteField("language", cfg.Language)
	}
	if err := mw.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.URL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", mw.FormDataContentType())

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return "", fmt.Errorf("transcribe failed: %s: %s", res.Status, b)
	}

	var out response
	if err := json.Unmarshal(b, &out); err != nil {
		return "", err
	}
	if out.Text == "" {
		return "", fmt.Errorf("empty transcript response: %s", b)
	}
	return out.Text, nil
}
