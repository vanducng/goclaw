package cmd

import (
	"context"
	"log/slog"

	"github.com/nextlevelbuilder/goclaw/internal/store"
	"github.com/nextlevelbuilder/goclaw/internal/tools"
)

// builtinToolSeedData returns the canonical list of built-in tools to seed into the database.
// Seed preserves user-customized enabled/settings values across upgrades.
func builtinToolSeedData() []store.BuiltinToolDef {
	return []store.BuiltinToolDef{
		// filesystem
		{Name: "read_file", DisplayName: "Read File", Description: "Read file contents from the workspace", Category: "filesystem", Enabled: true},
		{Name: "write_file", DisplayName: "Write File", Description: "Write or create files in the workspace", Category: "filesystem", Enabled: true},
		{Name: "list_files", DisplayName: "List Files", Description: "List files and directories in the workspace", Category: "filesystem", Enabled: true},
		{Name: "edit", DisplayName: "Edit File", Description: "Apply targeted edits to files (search and replace)", Category: "filesystem", Enabled: true},

		// runtime
		{Name: "exec", DisplayName: "Execute Command", Description: "Execute shell commands in the workspace", Category: "runtime", Enabled: true},

		// web
		{Name: "web_search", DisplayName: "Web Search", Description: "Search the web using Brave or DuckDuckGo", Category: "web", Enabled: true},
		{Name: "web_fetch", DisplayName: "Web Fetch", Description: "Fetch and extract content from web URLs", Category: "web", Enabled: true},

		// memory
		{Name: "memory_search", DisplayName: "Memory Search", Description: "Search through stored memory entries", Category: "memory", Enabled: true},
		{Name: "memory_get", DisplayName: "Memory Get", Description: "Retrieve a specific memory entry by key", Category: "memory", Enabled: true},

		// media
		{Name: "read_image", DisplayName: "Read Image", Description: "Analyze images using a vision-capable LLM provider", Category: "media", Enabled: true},
		{Name: "create_image", DisplayName: "Create Image", Description: "Generate images from text prompts using an image generation provider", Category: "media", Enabled: true},
		{Name: "tts", DisplayName: "Text to Speech", Description: "Convert text to speech audio", Category: "media", Enabled: true},

		// browser
		{Name: "browser", DisplayName: "Browser", Description: "Automate browser interactions (navigate, click, screenshot)", Category: "browser", Enabled: true},

		// sessions
		{Name: "sessions_list", DisplayName: "List Sessions", Description: "List active chat sessions", Category: "sessions", Enabled: true},
		{Name: "session_status", DisplayName: "Session Status", Description: "Get status of a chat session", Category: "sessions", Enabled: true},
		{Name: "sessions_history", DisplayName: "Session History", Description: "Get message history of a chat session", Category: "sessions", Enabled: true},
		{Name: "sessions_send", DisplayName: "Send to Session", Description: "Send a message to a chat session", Category: "sessions", Enabled: true},

		// messaging
		{Name: "message", DisplayName: "Message", Description: "Send messages to connected channels (Telegram, Discord, etc.)", Category: "messaging", Enabled: true},

		// scheduling
		{Name: "cron", DisplayName: "Cron Scheduler", Description: "Schedule recurring tasks with cron expressions", Category: "scheduling", Enabled: true},

		// subagents
		{Name: "spawn", DisplayName: "Spawn Subagent", Description: "Spawn an asynchronous background subagent", Category: "subagents", Enabled: true},
		{Name: "subagent", DisplayName: "Subagent", Description: "Run a synchronous subagent and wait for result", Category: "subagents", Enabled: true},

		// skills
		{Name: "skill_search", DisplayName: "Skill Search", Description: "Search available skills by keyword or description", Category: "skills", Enabled: true},

		// delegation
		{Name: "delegate", DisplayName: "Delegate", Description: "Delegate a task to another agent", Category: "delegation", Enabled: true},
		{Name: "delegate_search", DisplayName: "Delegate Search", Description: "Search for agents to delegate tasks to", Category: "delegation", Enabled: true},
		{Name: "evaluate_loop", DisplayName: "Evaluate Loop", Description: "Run an evaluate-optimize loop with delegated agents", Category: "delegation", Enabled: true},
		{Name: "handoff", DisplayName: "Handoff", Description: "Transfer conversation to another agent", Category: "delegation", Enabled: true},

		// teams
		{Name: "team_tasks", DisplayName: "Team Tasks", Description: "Manage tasks within a team of agents", Category: "teams", Enabled: true},
		{Name: "team_message", DisplayName: "Team Message", Description: "Send messages between team agents", Category: "teams", Enabled: true},
	}
}

// seedBuiltinTools seeds built-in tool definitions into the database.
// Idempotent: preserves user-customized enabled/settings on conflict.
func seedBuiltinTools(ctx context.Context, bts store.BuiltinToolStore) {
	seeds := builtinToolSeedData()
	if err := bts.Seed(ctx, seeds); err != nil {
		slog.Error("failed to seed builtin tools", "error", err)
		return
	}
	slog.Info("builtin tools seeded", "count", len(seeds))
}

// applyBuiltinToolDisables unregisters disabled builtin tools from the registry.
// Called at startup and on cache invalidation.
func applyBuiltinToolDisables(ctx context.Context, bts store.BuiltinToolStore, toolsReg *tools.Registry) {
	all, err := bts.List(ctx)
	if err != nil {
		slog.Warn("failed to list builtin tools for disable check", "error", err)
		return
	}

	var disabled int
	for _, t := range all {
		if !t.Enabled {
			toolsReg.Unregister(t.Name)
			disabled++
		}
	}
	if disabled > 0 {
		slog.Info("builtin tools disabled", "count", disabled)
	}
}
