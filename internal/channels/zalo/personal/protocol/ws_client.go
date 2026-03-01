package protocol

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/coder/websocket"
)

// WSClient wraps coder/websocket with a thread-safe write method.
type WSClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// DialWS connects to a WebSocket endpoint using coder/websocket.
// No compression is negotiated — avoids RSV1 issues with Zalo's servers.
func DialWS(ctx context.Context, wsURL string, headers http.Header, jar http.CookieJar) (*WSClient, error) {
	opts := &websocket.DialOptions{
		HTTPHeader: headers,
		HTTPClient: &http.Client{Jar: jar},
	}

	conn, _, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		return nil, fmt.Errorf("zalo_personal: ws dial: %w", err)
	}
	conn.SetReadLimit(1 << 20) // 1MB
	return &WSClient{conn: conn}, nil
}

// ReadMessage reads the next WebSocket message. Blocks until a message
// arrives, the context is cancelled, or the connection is closed.
func (c *WSClient) ReadMessage(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.Read(ctx)
	return data, err
}

// WriteMessage sends a binary WebSocket message. Thread-safe.
func (c *WSClient) WriteMessage(ctx context.Context, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.Write(ctx, websocket.MessageBinary, data)
}

// Close sends a close frame and shuts down the connection.
func (c *WSClient) Close(code int, reason string) {
	c.conn.Close(websocket.StatusCode(code), reason)
}

// parseWSCloseInfo extracts close code and reason from a coder/websocket error.
func parseWSCloseInfo(err error) CloseInfo {
	var ce websocket.CloseError
	if errors.As(err, &ce) {
		return CloseInfo{Code: int(ce.Code), Reason: ce.Reason}
	}
	return CloseInfo{Code: 1006, Reason: err.Error()}
}
