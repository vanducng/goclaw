package protocol

import "encoding/json"

// Message is the interface for incoming Zalo messages (DM or group).
type Message interface {
	Type() ThreadType
	ThreadID() string
	IsSelf() bool
}

// TMessage is the raw JSON message payload from Zalo WebSocket.
type TMessage struct {
	MsgID   string  `json:"msgId"`
	UIDFrom string  `json:"uidFrom"`
	IDTo    string  `json:"idTo"`
	DName   string  `json:"dName"`
	TS      string  `json:"ts"`
	Content Content `json:"content"`
	MsgType string  `json:"msgType"`
	CMD     int     `json:"cmd"`
	ST      int     `json:"st"`
	AT      int     `json:"at"`
}

// TGroupMessage extends TMessage with group-specific fields.
type TGroupMessage struct {
	TMessage
	Mentions []*TMention `json:"mentions,omitempty"`
}

// TMention represents an @mention in a group message.
type TMention struct {
	UID  string      `json:"uid"`  // user ID or "-1" for @all
	Pos  int         `json:"pos"`
	Len  int         `json:"len"`
	Type MentionType `json:"type"` // 0=individual, 1=all
}

// MentionType distinguishes individual vs @all mentions.
type MentionType int

const (
	MentionEach MentionType = 0
	MentionAll  MentionType = 1
	MentionAllUID           = "-1"
)

// Content is a union type: can be a plain string or an attachment object.
// For MVP, we only extract the string value.
type Content struct {
	String *string
}

func (c *Content) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		c.String = &s
		return nil
	}
	// Non-string content (attachments, stickers, etc.) — ignore for MVP
	return nil
}

func (c Content) MarshalJSON() ([]byte, error) {
	if c.String != nil {
		return json.Marshal(c.String)
	}
	return []byte("null"), nil
}

// Text returns the plain text content, or empty string for non-text.
func (c Content) Text() string {
	if c.String != nil {
		return *c.String
	}
	return ""
}

// UserMessage represents a DM (type=0).
type UserMessage struct {
	Data     TMessage
	threadID string
	isSelf   bool
}

// NewUserMessage creates a UserMessage, resolving self-sent messages.
func NewUserMessage(selfUID string, data TMessage) UserMessage {
	msg := UserMessage{Data: data, threadID: data.UIDFrom}
	msg.isSelf = data.UIDFrom == DefaultUIDSelf

	if data.UIDFrom == DefaultUIDSelf {
		msg.threadID = data.IDTo
		msg.Data.UIDFrom = selfUID
	}
	if data.IDTo == DefaultUIDSelf {
		msg.Data.IDTo = selfUID
	}
	return msg
}

func (m UserMessage) Type() ThreadType { return ThreadTypeUser }
func (m UserMessage) ThreadID() string { return m.threadID }
func (m UserMessage) IsSelf() bool     { return m.isSelf }

// GroupMessage represents a group message (type=1).
type GroupMessage struct {
	Data     TGroupMessage
	threadID string
	isSelf   bool
}

// NewGroupMessage creates a GroupMessage, resolving self-sent messages.
func NewGroupMessage(selfUID string, data TGroupMessage) GroupMessage {
	g := GroupMessage{Data: data, threadID: data.IDTo}
	g.isSelf = data.UIDFrom == DefaultUIDSelf
	if data.UIDFrom == DefaultUIDSelf {
		g.Data.UIDFrom = selfUID
	}
	return g
}

func (m GroupMessage) Type() ThreadType { return ThreadTypeGroup }
func (m GroupMessage) ThreadID() string { return m.threadID }
func (m GroupMessage) IsSelf() bool     { return m.isSelf }
