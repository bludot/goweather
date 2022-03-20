package http_client

import "net/http"

// HTTPClient interface
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	Client HTTPClient
}

func NewClient(httpclient HTTPClient) HTTPClient {
	return &Client{
		Client: httpclient,
	}
}

func (c Client) Do(req *http.Request) (*http.Response, error) {
	return c.Client.Do(req)
}
