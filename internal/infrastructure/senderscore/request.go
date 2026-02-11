package senderscore

import (
	"fmt"
	"io"
	"net/http"
)

var baseUrl = "https://senderscore.org"

type HttpClientInterface interface {
	Do(req *http.Request) (*http.Response, error)
}

type RequestWrapper struct {
	client HttpClientInterface
}

func NewRequestWrapper(client HttpClientInterface) *RequestWrapper {
	return &RequestWrapper{client: client}
}

func (r *RequestWrapper) SendRequest(method, url string, userAgent string) ([]byte, error) {
	req, err := http.NewRequest(method, r.buildUrl(url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server error: status %d, body %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (r *RequestWrapper) buildUrl(url string) string {
	return fmt.Sprintf("%s/%s", baseUrl, url)
}
