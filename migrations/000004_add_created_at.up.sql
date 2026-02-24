-- Add created_at to tables that only have updated_at (or no timestamp at all).
-- Existing rows get NOW() as default (close enough for backfill).

ALTER TABLE agent_context_files ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE config_secrets ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE cron_run_logs ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE embedding_cache ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE memory_chunks ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE memory_documents ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE user_context_files ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE user_agent_overrides ADD COLUMN created_at TIMESTAMPTZ DEFAULT NOW();
ALTER TABLE user_agent_overrides ADD COLUMN updated_at TIMESTAMPTZ DEFAULT NOW();
