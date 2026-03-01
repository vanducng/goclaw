package protocol

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	CloseCodeDuplicate = 3000 // another Zalo session opened — never reconnect
	msgBufferSize      = 64
	minEncDataLen      = 48
)

// Listener connects to Zalo's WebSocket and dispatches messages.
type Listener struct {
	mu   sync.RWMutex
	sess *Session

	wsURLs      []string
	wsURL       string
	rotateCount int

	conn      *websocket.Conn
	cipherKey string

	retryStates map[string]*retryState

	messageCh      chan Message
	disconnectedCh chan CloseInfo
	closedCh       chan CloseInfo
	errorCh        chan error

	pingCancel context.CancelFunc
	cancel     context.CancelFunc
	wg         sync.WaitGroup
}

// CloseInfo carries the WebSocket close code and reason.
type CloseInfo struct {
	Code   int
	Reason string
}

type retryState struct {
	count int
	max   int
	times []int // delay in ms per retry attempt
}

// NewListener creates a listener from an authenticated session.
func NewListener(sess *Session) (*Listener, error) {
	if sess.LoginInfo == nil || len(sess.LoginInfo.ZpwWebsocket) == 0 {
		return nil, fmt.Errorf("zca: no websocket URLs in session")
	}

	wsURL := buildWSURL(sess, sess.LoginInfo.ZpwWebsocket[0])
	return &Listener{
		sess:           sess,
		wsURLs:         sess.LoginInfo.ZpwWebsocket,
		wsURL:          wsURL,
		retryStates:    buildListenerRetryStates(sess.Settings),
		messageCh:      make(chan Message, msgBufferSize),
		disconnectedCh: make(chan CloseInfo, 4),
		closedCh:       make(chan CloseInfo, 1),
		errorCh:        make(chan error, 16),
	}, nil
}

// Channel accessors.
func (ln *Listener) Messages() <-chan Message      { return ln.messageCh }
func (ln *Listener) Disconnected() <-chan CloseInfo { return ln.disconnectedCh }
func (ln *Listener) Closed() <-chan CloseInfo       { return ln.closedCh }
func (ln *Listener) Errors() <-chan error           { return ln.errorCh }

// Start connects to WebSocket and begins reading messages.
func (ln *Listener) Start(ctx context.Context) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.conn != nil {
		return fmt.Errorf("zca: listener already started")
	}

	lctx, cancel := context.WithCancel(ctx)
	ln.cancel = cancel

	u, _ := url.Parse(ln.wsURL)
	h := http.Header{}
	h.Set("Accept-Encoding", "gzip")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("Cache-Control", "no-cache")
	h.Set("Host", u.Host)
	h.Set("Origin", DefaultBaseURL.String())
	h.Set("Pragma", "no-cache")
	h.Set("User-Agent", ln.sess.UserAgent)

	conn, _, err := websocket.Dial(lctx, ln.wsURL, &websocket.DialOptions{
		HTTPHeader: h,
		HTTPClient: ln.sess.Client,
	})
	if err != nil {
		cancel()
		return fmt.Errorf("zca: ws dial: %w", err)
	}
	conn.SetReadLimit(1 << 20) // 1MB

	ln.conn = conn
	ln.wg.Add(1)
	go ln.run(lctx)
	return nil
}

// Stop gracefully closes the WebSocket connection.
func (ln *Listener) Stop() {
	ln.mu.Lock()
	conn := ln.conn
	cancel := ln.cancel
	ln.mu.Unlock()

	if conn == nil {
		return
	}
	if cancel != nil {
		cancel()
	}
	conn.Close(websocket.StatusNormalClosure, "")
	ln.wg.Wait()
}

func (ln *Listener) run(ctx context.Context) {
	defer ln.wg.Done()

	for {
		if ctx.Err() != nil {
			return
		}
		_, data, err := ln.conn.Read(ctx)
		if err != nil {
			ci := parseWSCloseInfo(err)
			ln.handleDisconnect(ctx, ci)
			return
		}
		ln.handleFrame(ctx, data)
	}
}

func (ln *Listener) handleFrame(ctx context.Context, data []byte) {
	if len(data) < 4 {
		return
	}

	version := data[0]
	cmd := binary.LittleEndian.Uint16(data[1:3])
	subCmd := data[3]
	body := data[4:]

	var envelope struct {
		Key     *string `json:"key"`
		Encrypt uint    `json:"encrypt"`
		Data    string  `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		emit(ctx, ln.errorCh, fmt.Errorf("zca: parse ws frame: %w", err))
		return
	}

	key := fmt.Sprintf("%d_%d_%d", version, cmd, subCmd)
	switch key {
	case "1_1_1":
		ln.handleCipherKey(ctx, envelope.Key)
	case "1_501_0":
		ln.handleUserMessages(ctx, envelope.Data, envelope.Encrypt)
	case "1_521_0":
		ln.handleGroupMessages(ctx, envelope.Data, envelope.Encrypt)
	case "1_3000_0":
		slog.Warn("zca: duplicate connection detected, closing")
		ln.mu.RLock()
		conn := ln.conn
		ln.mu.RUnlock()
		if conn != nil {
			conn.Close(websocket.StatusCode(CloseCodeDuplicate), "duplicate")
		}
	}
}

func (ln *Listener) handleCipherKey(ctx context.Context, key *string) {
	if key == nil || *key == "" {
		return
	}
	ln.mu.Lock()
	ln.cipherKey = *key
	ln.mu.Unlock()

	// Start ping loop
	if ln.sess.Settings != nil {
		interval := ln.sess.Settings.Features.Socket.PingInterval
		if interval > 0 {
			pctx, pcancel := context.WithCancel(ctx)
			ln.mu.Lock()
			ln.pingCancel = pcancel
			ln.mu.Unlock()
			go ln.pingLoop(pctx, time.Duration(interval)*time.Millisecond)
		}
	}
}
