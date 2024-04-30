package harness

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var _ Client = (*HTTPClient)(nil)

const (
	RestoreEndpoint = "/cache/intel/download?accountId=%s&cacheKey=%s"
	StoreEndpoint   = "/cache/intel/upload?accountId=%s&cacheKey=%s"
	ExistsEndpoint  = "/cache/intel/exists?accountId=%s&cacheKey=%s"
	ListEndpoint    = "/cache/intel/list?accountId=%s&cacheKey=%s"
)

// NewHTTPClient returns a new HTTPClient.
func New(endpoint, accountID, bearerToken string, skipverify bool) *HTTPClient {
	endpoint = strings.TrimSuffix(endpoint, "/")
	client := &HTTPClient{
		Endpoint:    endpoint,
		BearerToken: bearerToken,
		AccountID:   accountID,
		Client: &http.Client{
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
	return client
}

// HTTPClient provides an http service client.
type HTTPClient struct {
	Client      *http.Client
	Endpoint    string
	AccountID   string
	BearerToken string
}

// getUploadURL will get the 'put' presigned url from cache service
func (c *HTTPClient) GetUploadURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(StoreEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

// getDownloadURL will get the 'get' presigned url from cache service
func (c *HTTPClient) GetDownloadURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(RestoreEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

// getExistsURL will get the 'exists' presigned url from cache service
func (c *HTTPClient) GetExistsURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(ExistsEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

// getListURL will get the 'list' presigned url from cache service
func (c *HTTPClient) GetListURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(ListEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

func (c *HTTPClient) getLink(ctx context.Context, path string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return "", err
	}
	if c.BearerToken != "" {
		req.Header.Add("X-Harness-Token", c.BearerToken)
	}

	resp, err := c.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get link with status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *HTTPClient) client() *http.Client {
	return c.Client
}
