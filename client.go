package langdock

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type Config struct {
	// The API key used to authenticate requests.
	// This is required.
	APIKey string

	// The base URL for the Langdock Knowledge API.
	// The default value is "https://api.langdock.com".
	BaseURL string

	// The maximum number of retries for failed requests.
	// The default value is 3.
	MaxRetries int

	// The HTTP client to use for requests.
	// If nil, a default client with a 60-second timeout will be used.
	HTTPClient *http.Client
}

type Client struct {
	Config     Config
	HTTPClient *http.Client

	Knowledge *KnowledgeService
}

func New(opts ...func(*Config)) *Client {
	config := Config{
		APIKey:     "",
		BaseURL:    "https://api.langdock.com",
		MaxRetries: 3,
	}

	for _, configure := range opts {
		configure(&config)
	}

	client := &Client{
		Config: config,
	}

	if config.HTTPClient != nil {
		client.HTTPClient = config.HTTPClient
	} else {
		client.HTTPClient = &http.Client{
			Timeout: 60 * time.Second,
		}
	}

	client.Knowledge = NewKnowledgeService(client)

	return client
}

func WithAPIKey(apiKey string) func(*Config) {
	return func(c *Config) {
		c.APIKey = apiKey
	}
}

func WithBaseURL(baseURL string) func(*Config) {
	return func(c *Config) {
		c.BaseURL = baseURL
	}
}

func WithMaxRetries(maxRetries int) func(*Config) {
	return func(c *Config) {
		c.MaxRetries = maxRetries
	}
}

func WithHTTPClient(httpClient *http.Client) func(*Config) {
	return func(c *Config) {
		c.HTTPClient = httpClient
	}
}

func (c *Client) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	url := c.Config.BaseURL + path

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Config.APIKey))

	return req, nil
}

func (c *Client) Do(req *http.Request, v any) error {
	var last error

	for attempt := 0; attempt < c.Config.MaxRetries; attempt++ {
		if attempt > 0 {
			max := 5 * time.Second

			base := min(time.Duration(1<<uint(attempt-1))*100*time.Millisecond, max)

			jitter := time.Duration(rand.Int63n(int64(base))) - base/2
			wait := base + jitter
			if wait < 0 {
				wait = 0
			}

			select {
			case <-req.Context().Done():
				return req.Context().Err()
			case <-time.After(wait):
			}

			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return err
				}
				req.Body = body
			}
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			last = err
			continue
		}

		defer resp.Body.Close()

		if resp.StatusCode == http.StatusTooManyRequests {
			last = &RateLimitError{}
			continue
		}

		if resp.StatusCode >= 500 {
			last = fmt.Errorf("The request failed with status code %d.", resp.StatusCode)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			last = err
			continue
		}

		fmt.Println(string(body))

		if v != nil && len(body) > 0 {
			if err := json.Unmarshal(body, v); err != nil {
				last = err
				return last
			}
		}

		return nil
	}

	return last
}

type RateLimitError struct{}

func (e *RateLimitError) Error() string {
	return "Rate limit exceeded."
}
