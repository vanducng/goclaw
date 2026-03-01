package personal

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/nextlevelbuilder/goclaw/internal/bus"
	"github.com/nextlevelbuilder/goclaw/internal/channels"
	"github.com/nextlevelbuilder/goclaw/internal/channels/zalo/personal/protocol"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

const maxTextLength = 2000

// Channel connects to Zalo Personal Chat via the internal protocol port (from zcago, MIT).
// WARNING: This is an unofficial, reverse-engineered integration. Account may be locked/banned.
type Channel struct {
	*channels.BaseChannel
	config          config.ZaloPersonalConfig
	pairingService  store.PairingStore
	pairingDebounce sync.Map // senderID -> time.Time

	sess     *protocol.Session
	listener *protocol.Listener

	// Pre-loaded credentials (managed mode: from DB, standalone: from file or QR).
	preloadedCreds *protocol.Credentials

	requireMention bool
	stopCh         chan struct{}
	stopOnce       sync.Once
}

// New creates a new ZCA channel from config.
func New(cfg config.ZaloPersonalConfig, msgBus *bus.MessageBus, pairingSvc store.PairingStore) (*Channel, error) {
	base := channels.NewBaseChannel("zalo_personal", msgBus, cfg.AllowFrom)

	dmPolicy := cfg.DMPolicy
	if dmPolicy == "" {
		dmPolicy = "pairing"
	}
	groupPolicy := cfg.GroupPolicy
	if groupPolicy == "" {
		groupPolicy = "open"
	}
	base.ValidatePolicy(dmPolicy, groupPolicy)

	requireMention := true
	if cfg.RequireMention != nil {
		requireMention = *cfg.RequireMention
	}

	return &Channel{
		BaseChannel:    base,
		config:         cfg,
		pairingService: pairingSvc,
		requireMention: requireMention,
		stopCh:         make(chan struct{}),
	}, nil
}

// Start authenticates and begins listening for Zalo messages.
func (c *Channel) Start(ctx context.Context) error {
	slog.Warn("security.unofficial_api",
		"channel", "zalo_personal",
		"msg", "ZCA is unofficial and reverse-engineered. Account may be locked/banned. Use at own risk.",
	)

	sess, err := c.authenticate(ctx)
	if err != nil {
		return fmt.Errorf("zca auth: %w", err)
	}
	c.sess = sess

	slog.Info("zca connected", "uid", sess.UID)

	ln, err := protocol.NewListener(sess)
	if err != nil {
		return fmt.Errorf("zca listener: %w", err)
	}
	c.listener = ln

	if err := ln.Start(ctx); err != nil {
		return fmt.Errorf("zca listener start: %w", err)
	}

	c.SetRunning(true)
	go c.listenLoop(ctx)

	slog.Info("zca listener loop started")
	return nil
}

// Stop gracefully shuts down the ZCA channel.
func (c *Channel) Stop(_ context.Context) error {
	slog.Info("stopping zca channel")
	c.stopOnce.Do(func() { close(c.stopCh) })
	if c.listener != nil {
		c.listener.Stop()
	}
	c.SetRunning(false)
	return nil
}

// Send delivers an outbound message to a Zalo chat.
func (c *Channel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	if !c.IsRunning() || c.sess == nil {
		return fmt.Errorf("zca channel not running")
	}

	threadType := protocol.ThreadTypeUser
	if msg.Metadata != nil {
		if _, ok := msg.Metadata["group_id"]; ok {
			threadType = protocol.ThreadTypeGroup
		}
	}

	return c.sendChunkedText(ctx, msg.ChatID, threadType, msg.Content)
}

func (c *Channel) listenLoop(ctx context.Context) {
	defer c.SetRunning(false)

	for {
		select {
		case <-ctx.Done():
			slog.Info("zca listener loop stopped (context)")
			return
		case <-c.stopCh:
			slog.Info("zca listener loop stopped")
			return

		case msg, ok := <-c.listener.Messages():
			if !ok {
				return
			}
			c.handleMessage(msg)

		case ci := <-c.listener.Disconnected():
			slog.Warn("zca disconnected", "code", ci.Code, "reason", ci.Reason)

		case ci := <-c.listener.Closed():
			if ci.Code == protocol.CloseCodeDuplicate {
				slog.Warn("zca duplicate connection detected (code 3000). Close other Zalo sessions and restart.",
					"channel", c.Name())
			} else {
				slog.Warn("zca connection closed", "code", ci.Code, "reason", ci.Reason)
			}
			return

		case err := <-c.listener.Errors():
			slog.Warn("zca listener error", "error", err)
		}
	}
}

func (c *Channel) handleMessage(msg protocol.Message) {
	if msg.IsSelf() {
		return
	}

	switch m := msg.(type) {
	case protocol.UserMessage:
		c.handleDM(m)
	case protocol.GroupMessage:
		c.handleGroupMessage(m)
	}
}

func (c *Channel) handleDM(msg protocol.UserMessage) {
	senderID := msg.Data.UIDFrom
	threadID := msg.ThreadID()
	content := msg.Data.Content.Text()
	if content == "" {
		return
	}

	if !c.checkDMPolicy(senderID, threadID) {
		return
	}

	slog.Debug("zca DM received",
		"sender", senderID,
		"thread", threadID,
		"preview", channels.Truncate(content, 50),
	)

	metadata := map[string]string{
		"message_id": msg.Data.MsgID,
		"platform":   "zalo_personal",
	}
	c.HandleMessage(senderID, threadID, content, nil, metadata, "direct")
}

func (c *Channel) handleGroupMessage(msg protocol.GroupMessage) {
	senderID := msg.Data.UIDFrom
	threadID := msg.ThreadID()
	content := msg.Data.Content.Text()
	if content == "" {
		return
	}

	if !c.checkGroupPolicy(senderID, threadID, msg.Data.Mentions) {
		return
	}

	slog.Debug("zca group message received",
		"sender", senderID,
		"group", threadID,
		"preview", channels.Truncate(content, 50),
	)

	metadata := map[string]string{
		"message_id": msg.Data.MsgID,
		"platform":   "zalo_personal",
		"group_id":   threadID,
	}
	c.HandleMessage(senderID, threadID, content, nil, metadata, "group")
}

func (c *Channel) sendChunkedText(ctx context.Context, chatID string, threadType protocol.ThreadType, text string) error {
	for len(text) > 0 {
		chunk := text
		if len(chunk) > maxTextLength {
			cutAt := maxTextLength
			if idx := strings.LastIndex(text[:maxTextLength], "\n"); idx > maxTextLength/2 {
				cutAt = idx + 1
			}
			chunk = text[:cutAt]
			text = text[cutAt:]
		} else {
			text = ""
		}

		if _, err := protocol.SendMessage(ctx, c.sess, chatID, threadType, chunk); err != nil {
			return err
		}
	}
	return nil
}
