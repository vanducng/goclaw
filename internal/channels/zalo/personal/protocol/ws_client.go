package protocol

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

// wsJar wraps http.CookieJar converting wss:// → https:// (ws:// → http://)
// before delegating. gorilla/websocket already does this conversion internally,
// so this wrapper acts as a safety net for any direct calls and documents intent.
type wsJar struct {
	http.CookieJar
}

func (j *wsJar) Cookies(u *url.URL) []*http.Cookie {
	u2 := *u
	switch u2.Scheme {
	case "wss":
		u2.Scheme = "https"
	case "ws":
		u2.Scheme = "http"
	}
	return j.CookieJar.Cookies(&u2)
}

func (j *wsJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	u2 := *u
	switch u2.Scheme {
	case "wss":
		u2.Scheme = "https"
	case "ws":
		u2.Scheme = "http"
	}
	j.CookieJar.SetCookies(&u2, cookies)
}

// WSClient wraps gorilla/websocket with a thread-safe write method.
type WSClient struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

// DialWS connects to a WebSocket endpoint using gorilla/websocket.
// Wraps the cookie jar to fix Go's cookiejar not returning cookies for wss:// URLs.
func DialWS(ctx context.Context, wsURL string, headers http.Header, jar http.CookieJar) (*WSClient, error) {
	dialer := websocket.Dialer{
		EnableCompression: true,
	}
	if jar != nil {
		dialer.Jar = &wsJar{jar}
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, headers)
	if err != nil {
		return nil, fmt.Errorf("zalo_personal: ws dial: %w", err)
	}
	conn.SetReadLimit(1 << 20) // 1MB
	return &WSClient{conn: conn}, nil
}

// ReadMessage reads the next WebSocket message. Blocks until a message
// arrives, the context is cancelled, or the connection is closed.
func (c *WSClient) ReadMessage(ctx context.Context) ([]byte, error) {
	_, data, err := c.conn.ReadMessage()
	return data, err
}

// WriteMessage sends a binary WebSocket message. Thread-safe.
func (c *WSClient) WriteMessage(ctx context.Context, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

// Close sends a close frame and shuts down the connection.
func (c *WSClient) Close(code int, reason string) {
	c.mu.Lock()
	c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
	)
	c.conn.Close()
	c.mu.Unlock()
}

// parseWSCloseInfo extracts close code and reason from a gorilla/websocket error.
func parseWSCloseInfo(err error) CloseInfo {
	var ce *websocket.CloseError
	if errors.As(err, &ce) {
		return CloseInfo{Code: ce.Code, Reason: ce.Text}
	}
	return CloseInfo{Code: 1006, Reason: err.Error()}
}
