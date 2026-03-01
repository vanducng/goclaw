package protocol

import (
	"encoding/json"
	"testing"
)

func TestContent_UnmarshalJSON_String(t *testing.T) {
	var c Content
	if err := json.Unmarshal([]byte(`"hello world"`), &c); err != nil {
		t.Fatal(err)
	}
	if c.String == nil || *c.String != "hello world" {
		t.Errorf("expected 'hello world', got %v", c.String)
	}
	if c.Text() != "hello world" {
		t.Errorf("Text() = %q, want 'hello world'", c.Text())
	}
}

func TestContent_UnmarshalJSON_Object(t *testing.T) {
	// Non-string content (attachment) should not error, just be nil
	var c Content
	if err := json.Unmarshal([]byte(`{"type":"sticker","id":123}`), &c); err != nil {
		t.Fatal(err)
	}
	if c.String != nil {
		t.Errorf("expected nil String for object content, got %q", *c.String)
	}
	if c.Text() != "" {
		t.Errorf("Text() = %q, want empty", c.Text())
	}
}

func TestContent_MarshalJSON(t *testing.T) {
	s := "test message"
	c := Content{String: &s}
	b, err := c.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != `"test message"` {
		t.Errorf("got %s, want %q", b, `"test message"`)
	}

	// Nil content marshals as null
	c2 := Content{}
	b2, err := c2.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b2) != "null" {
		t.Errorf("got %s, want null", b2)
	}
}

func TestNewUserMessage(t *testing.T) {
	selfUID := "12345"
	text := "hello"

	t.Run("incoming message", func(t *testing.T) {
		msg := TMessage{
			MsgID:   "m1",
			UIDFrom: "67890",
			IDTo:    selfUID,
			Content: Content{String: &text},
		}
		um := NewUserMessage(selfUID, msg)

		if um.Type() != ThreadTypeUser {
			t.Errorf("Type = %d, want %d", um.Type(), ThreadTypeUser)
		}
		if um.ThreadID() != "67890" {
			t.Errorf("ThreadID = %q, want '67890'", um.ThreadID())
		}
		if um.IsSelf() {
			t.Error("IsSelf should be false for incoming message")
		}
	})

	t.Run("self-sent message", func(t *testing.T) {
		msg := TMessage{
			MsgID:   "m2",
			UIDFrom: DefaultUIDSelf,
			IDTo:    "67890",
			Content: Content{String: &text},
		}
		um := NewUserMessage(selfUID, msg)

		if !um.IsSelf() {
			t.Error("IsSelf should be true for self-sent")
		}
		if um.ThreadID() != "67890" {
			t.Errorf("ThreadID = %q, want '67890' (should use IDTo for self-sent)", um.ThreadID())
		}
		if um.Data.UIDFrom != selfUID {
			t.Errorf("UIDFrom = %q, want %q (should resolve '0' to selfUID)", um.Data.UIDFrom, selfUID)
		}
	})

	t.Run("IDTo is self", func(t *testing.T) {
		msg := TMessage{
			MsgID:   "m3",
			UIDFrom: "67890",
			IDTo:    DefaultUIDSelf,
		}
		um := NewUserMessage(selfUID, msg)

		if um.Data.IDTo != selfUID {
			t.Errorf("IDTo = %q, want %q (should resolve '0' to selfUID)", um.Data.IDTo, selfUID)
		}
	})
}

func TestNewGroupMessage(t *testing.T) {
	selfUID := "12345"
	text := "group msg"

	t.Run("incoming group message", func(t *testing.T) {
		msg := TGroupMessage{
			TMessage: TMessage{
				MsgID:   "gm1",
				UIDFrom: "67890",
				IDTo:    "group_abc",
				Content: Content{String: &text},
			},
		}
		gm := NewGroupMessage(selfUID, msg)

		if gm.Type() != ThreadTypeGroup {
			t.Errorf("Type = %d, want %d", gm.Type(), ThreadTypeGroup)
		}
		if gm.ThreadID() != "group_abc" {
			t.Errorf("ThreadID = %q, want 'group_abc'", gm.ThreadID())
		}
		if gm.IsSelf() {
			t.Error("IsSelf should be false")
		}
	})

	t.Run("self-sent group message", func(t *testing.T) {
		msg := TGroupMessage{
			TMessage: TMessage{
				MsgID:   "gm2",
				UIDFrom: DefaultUIDSelf,
				IDTo:    "group_abc",
			},
		}
		gm := NewGroupMessage(selfUID, msg)

		if !gm.IsSelf() {
			t.Error("IsSelf should be true for self-sent")
		}
		if gm.Data.UIDFrom != selfUID {
			t.Errorf("UIDFrom = %q, want %q", gm.Data.UIDFrom, selfUID)
		}
		if gm.ThreadID() != "group_abc" {
			t.Errorf("ThreadID = %q, want 'group_abc'", gm.ThreadID())
		}
	})
}

func TestTMessage_UnmarshalJSON(t *testing.T) {
	raw := `{
		"msgId": "123",
		"uidFrom": "456",
		"idTo": "789",
		"dName": "Test User",
		"ts": "1709300000",
		"content": "hello",
		"msgType": "chat.message",
		"cmd": 501,
		"st": 1,
		"at": 0
	}`

	var msg TMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if msg.MsgID != "123" {
		t.Errorf("MsgID = %q", msg.MsgID)
	}
	if msg.Content.Text() != "hello" {
		t.Errorf("Content = %q", msg.Content.Text())
	}
	if msg.CMD != 501 {
		t.Errorf("CMD = %d", msg.CMD)
	}
}

func TestTGroupMessage_WithMentions(t *testing.T) {
	raw := `{
		"msgId": "gm1",
		"uidFrom": "111",
		"idTo": "group1",
		"content": "@all hello",
		"msgType": "chat.message",
		"cmd": 521,
		"st": 1,
		"at": 0,
		"mentions": [
			{"uid": "-1", "pos": 0, "len": 4, "type": 1}
		]
	}`

	var msg TGroupMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		t.Fatal(err)
	}
	if len(msg.Mentions) != 1 {
		t.Fatalf("expected 1 mention, got %d", len(msg.Mentions))
	}
	m := msg.Mentions[0]
	if m.UID != MentionAllUID {
		t.Errorf("mention UID = %q, want %q", m.UID, MentionAllUID)
	}
	if m.Type != MentionAll {
		t.Errorf("mention Type = %d, want %d", m.Type, MentionAll)
	}
}
