package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

// Client wraps NATS connection for request-reply and event subscriptions
type Client struct {
	nc *nats.Conn
}

// NewClient creates a new NATS client
func NewClient(nc *nats.Conn) *Client {
	return &Client{nc: nc}
}

// Request sends a request and waits for a reply
func (c *Client) Request(ctx context.Context, subject string, data []byte, timeout time.Duration) ([]byte, error) {
	// Create context with timeout if not already set
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	
	msg, err := c.nc.RequestWithContext(ctx, subject, data)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout: service did not respond within %v", timeout)
		}
		return nil, fmt.Errorf("failed to send request to %s: %w", subject, err)
	}
	return msg.Data, nil
}

// RequestJSON sends a JSON request and unmarshals the response
func (c *Client) RequestJSON(ctx context.Context, subject string, req interface{}, resp interface{}, timeout time.Duration) error {
	reqData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	respData, err := c.Request(ctx, subject, reqData, timeout)
	if err != nil {
		return err
	}

	// Check if response is an error
	var errorResp struct {
		Error   string `json:"error"`
		Details string `json:"details"`
	}
	if err := json.Unmarshal(respData, &errorResp); err == nil && errorResp.Error != "" {
		return fmt.Errorf("%s: %s", errorResp.Error, errorResp.Details)
	}

	if err := json.Unmarshal(respData, resp); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

// Subscribe subscribes to a subject and calls the handler for each message
func (c *Client) Subscribe(subject string, handler func(*nats.Msg)) (*nats.Subscription, error) {
	return c.nc.Subscribe(subject, handler)
}

// Close closes the NATS connection
func (c *Client) Close() {
	c.nc.Close()
}

