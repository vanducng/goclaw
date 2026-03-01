package personal

import (
	"encoding/json"
	"fmt"

	"github.com/nextlevelbuilder/goclaw/internal/bus"
	"github.com/nextlevelbuilder/goclaw/internal/channels"
	"github.com/nextlevelbuilder/goclaw/internal/channels/zalo/personal/protocol"
	"github.com/nextlevelbuilder/goclaw/internal/config"
	"github.com/nextlevelbuilder/goclaw/internal/store"
)

// zcaCreds maps the credentials JSON from the channel_instances table.
type zcaCreds struct {
	IMEI      string               `json:"imei"`
	Cookie    *protocol.CookieUnion `json:"cookie"`
	UserAgent string               `json:"userAgent"`
	Language  *string              `json:"language,omitempty"`
}

// zcaInstanceConfig maps the config JSONB from the channel_instances table.
type zcaInstanceConfig struct {
	DMPolicy       string   `json:"dm_policy,omitempty"`
	GroupPolicy    string   `json:"group_policy,omitempty"`
	RequireMention *bool    `json:"require_mention,omitempty"`
	AllowFrom      []string `json:"allow_from,omitempty"`
}

// Factory creates a ZCA channel from DB instance data (managed mode).
// Does NOT trigger QR login — credentials must be provided.
func Factory(name string, creds json.RawMessage, cfg json.RawMessage,
	msgBus *bus.MessageBus, pairingSvc store.PairingStore) (channels.Channel, error) {

	var c zcaCreds
	if len(creds) > 0 {
		if err := json.Unmarshal(creds, &c); err != nil {
			return nil, fmt.Errorf("decode zca credentials: %w", err)
		}
	}

	// Credentials required for managed mode (no interactive QR).
	if c.IMEI == "" || c.Cookie == nil {
		return nil, fmt.Errorf("zca credentials required (imei + cookie). Use QR login in standalone mode first, then export credentials")
	}

	var ic zcaInstanceConfig
	if len(cfg) > 0 {
		if err := json.Unmarshal(cfg, &ic); err != nil {
			return nil, fmt.Errorf("decode zca config: %w", err)
		}
	}

	zcaCfg := config.ZaloPersonalConfig{
		Enabled:        true,
		AllowFrom:      ic.AllowFrom,
		DMPolicy:       ic.DMPolicy,
		GroupPolicy:    ic.GroupPolicy,
		RequireMention: ic.RequireMention,
	}

	ch, err := New(zcaCfg, msgBus, pairingSvc)
	if err != nil {
		return nil, err
	}

	protoCred := &protocol.Credentials{
		IMEI:      c.IMEI,
		Cookie:    c.Cookie,
		UserAgent: c.UserAgent,
		Language:  c.Language,
	}
	ch.SetPreloadedCredentials(protoCred)
	ch.SetName(name)

	return ch, nil
}
