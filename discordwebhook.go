package discordwebhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type WebhookError struct {
	StatusCode int
	RetryAfter time.Duration
	Body       string
}

func (e *WebhookError) Error() string {
	if e.StatusCode == 429 {
		return fmt.Sprintf("rate limited: retry after %s", e.RetryAfter)
	}
	return fmt.Sprintf("discord webhook error %d: %s", e.StatusCode, e.Body)
}

func SendMessage(url string, message Message) error {
	payload := new(bytes.Buffer)
	if err := json.NewEncoder(payload).Encode(message); err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", payload)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 || resp.StatusCode == 204 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == 429 {
		var rateLimitResp struct {
			RetryAfter float64 `json:"retry_after"`
		}
		retryAfter := 0.0
		if jsonErr := json.Unmarshal(body, &rateLimitResp); jsonErr == nil {
			retryAfter = rateLimitResp.RetryAfter
		}
		return &WebhookError{
			StatusCode: 429,
			RetryAfter: time.Duration(retryAfter * float64(time.Second)),
			Body:       string(body),
		}
	}

	return &WebhookError{
		StatusCode: resp.StatusCode,
		RetryAfter: 0,
		Body:       string(body),
	}
}
