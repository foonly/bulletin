package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	client *http.Client
}

// NewApp creates a new App struct
func NewApp() *App {
	jar, _ := cookiejar.New(nil)
	return &App{
		client: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
	}
}

// Fetch handles HTTP requests from the frontend to bypass CORS
func (a *App) Fetch(method, url string, body string, headers map[string]string) (map[string]interface{}, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}

	// Set a default User-Agent
	req.Header.Set("User-Agent", "BulletinDesktop/1.0 (Wails)")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// Important: Handle cookies manually if necessary, or ensure the jar is working.
	// The http.Client with a Jar automatically handles cookies for subsequent requests
	// to the same domain.

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Convert headers to a simple map for JS
	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			respHeaders[k] = v[0]
		}
	}

	return map[string]interface{}{
		"status":  resp.StatusCode,
		"body":    string(respBody),
		"headers": respHeaders,
	}, nil
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ShowWindow shows the main window
func (a *App) ShowWindow() {
	runtime.WindowShow(a.ctx)
}

// HideWindow hides the main window
func (a *App) HideWindow() {
	runtime.WindowHide(a.ctx)
}
