package central

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jpillora/backoff"
	"go.uber.org/zap"
)

type Option func(c *Client) error

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		c.hc = client
		return nil
	}
}

func WithLogger(logger *zap.SugaredLogger) Option {
	return func(c *Client) error {
		c.logger = logger.Named("central")
		return nil
	}
}

func WithSessionKey(s string) Option {
	return func(c *Client) error {
		c.sessionKey = s
		return nil
	}
}

type Client struct {
	hc         *http.Client
	logger     *zap.SugaredLogger
	url        url.URL
	sessionKey string
}

func Open(serviceURL url.URL, opts ...Option) (*Client, error) {
	c := Client{
		hc:     http.DefaultClient,
		logger: zap.NewNop().Sugar(),
		url:    serviceURL,
	}
	for _, opt := range opts {
		if err := opt(&c); err != nil {
			return nil, err
		}
	}
	return &c, nil
}

func (c *Client) WithOpts(opts ...Option) (*Client, error) {
	newC := *c
	for _, opt := range opts {
		if err := opt(&newC); err != nil {
			return nil, err
		}
	}
	return &newC, nil
}

func (c *Client) GetMembershipsByIdentity(ctx context.Context, identityID int) ([]Membership, error) {
	var memberships []Membership
	_, err := c.doGET(ctx, fmt.Sprintf("/identities/%d/memberships", identityID), nil, &memberships)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return memberships, nil
}

func (c *Client) GetUserByIdentity(ctx context.Context, identityID int) (*User, error) {
	var user User
	_, err := c.doGET(ctx, fmt.Sprintf("/users/by-identity/%d", identityID), nil, &user)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (c *Client) GetApplicationByKey(ctx context.Context, key string) (*Application, error) {
	var application Application
	_, err := c.doGET(ctx, fmt.Sprintf("/applications/keys/%s", key), nil, &application)
	if err != nil {
		if isStatus(err, http.StatusNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &application, nil
}

func (c *Client) doGET(
	ctx context.Context,
	path string,
	params url.Values,
	output interface{}) (*http.Response, error) {
	req, err := c.newRequest(http.MethodGet, path, params)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	bckoff := &backoff.Backoff{Jitter: true}
	for {
		startTime := time.Now()
		resp, err := c.hc.Do(req)
		if err != nil {
			return nil, fmt.Errorf("GET request to %s failed: %w", req.URL, err)
		}
		defer func() {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()

		err, ok := c.checkResponse(req, resp, startTime)
		if ok {
			return resp, err
		}

		err = decodeResponseAsJSON(resp, resp.Body, output)
		if err != nil {
			_ = resp.Body.Close()
			c.logger.Warnf("Response error, will retry: %s", err)
			time.Sleep(bckoff.Duration())
			continue
		}
		return resp, nil
	}
}

func (c *Client) newRequest(method string, path string, params url.Values) (*http.Request, error) {
	url := c.formatURL(path, params)

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	if k := c.sessionKey; k != "" {
		req.Header.Set("Cookie", "checkpoint.session="+k)
	}

	return req, nil
}

func (c *Client) formatURL(path string, params url.Values) string {
	u := c.url
	u.Path = "/api/central/v1" + path
	if params != nil {
		u.RawQuery = params.Encode()
	}
	return u.String()
}

func (c *Client) checkResponse(
	req *http.Request,
	resp *http.Response,
	startTime time.Time) (error, bool) {
	c.logger.Infow(req.Method,
		"url", req.URL.String(),
		"time", time.Since(startTime).Seconds(),
		"status", resp.StatusCode)
	return errorFromResponse(req, resp, "Grove")
}
