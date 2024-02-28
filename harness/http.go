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
	harnessRestoreLinkEndpoint = "/cache/harness/download?accountId=%s&cacheKey=%s"
	harnessStoreLinkEndpoint   = "/cache/harness/upload?accountId=%s&cacheKey=%s"
	harnessExistsLinkEndpoint  = "/cache/harness/exists?accountId=%s&cacheKey=%s"
	harnessListLinkEndpoint    = "/cache/harness/list?accountId=%s&cacheKey=%s"
)

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

type HTTPClient struct {
	Client      *http.Client
	Endpoint    string
	AccountID   string
	BearerToken string
}

func (c *HTTPClient) GetUploadPresignURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(harnessStoreLinkEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

func (c *HTTPClient) GetDownloadPresignURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(harnessRestoreLinkEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

func (c *HTTPClient) GetExistsPresignURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(harnessExistsLinkEndpoint, c.AccountID, key)
	return c.getLink(ctx, c.Endpoint+path)
}

func (c *HTTPClient) GetListPresignURL(ctx context.Context, key string) (string, error) {
	path := fmt.Sprintf(harnessListLinkEndpoint, c.AccountID, key)
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
