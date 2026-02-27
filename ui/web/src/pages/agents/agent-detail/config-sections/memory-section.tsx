import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import type { MemoryConfig } from "@/types/agent";
import { ConfigSection, InfoLabel, numOrUndef } from "./config-section";

interface MemorySectionProps {
  enabled: boolean;
  value: MemoryConfig;
  onToggle: (v: boolean) => void;
  onChange: (v: MemoryConfig) => void;
}

export function MemorySection({ enabled, value, onToggle, onChange }: MemorySectionProps) {
  return (
    <ConfigSection
      title="Memory"
      description="Semantic memory search and embedding configuration"
      enabled={enabled}
      onToggle={onToggle}
    >
      <div className="flex items-center gap-2">
        <Switch
          checked={value.enabled ?? true}
          onCheckedChange={(v) => onChange({ ...value, enabled: v })}
        />
        <InfoLabel tip="Enable or disable the memory system for this agent. When enabled, the agent can store and recall information across sessions.">Enabled</InfoLabel>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <InfoLabel tip="LLM provider used for generating text embeddings. Leave empty to auto-detect from the agent's main provider.">Embedding Provider</InfoLabel>
          <Input
            placeholder="(auto)"
            value={value.embedding_provider ?? ""}
            onChange={(e) => onChange({ ...value, embedding_provider: e.target.value || undefined })}
          />
        </div>
        <div className="space-y-2">
          <InfoLabel tip="Embedding model name. Must be supported by the embedding provider (e.g. text-embedding-3-small for OpenAI).">Embedding Model</InfoLabel>
          <Input
            placeholder="text-embedding-3-small"
            value={value.embedding_model ?? ""}
            onChange={(e) => onChange({ ...value, embedding_model: e.target.value || undefined })}
          />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <InfoLabel tip="Maximum number of memory entries returned per search query.">Max Results</InfoLabel>
          <Input
            type="number"
            placeholder="6"
            value={value.max_results ?? ""}
            onChange={(e) => onChange({ ...value, max_results: numOrUndef(e.target.value) })}
          />
        </div>
        <div className="space-y-2">
          <InfoLabel tip="Maximum character length for each memory chunk. Longer content is split into smaller chunks before storing.">Max Chunk Length</InfoLabel>
          <Input
            type="number"
            placeholder="1000"
            value={value.max_chunk_len ?? ""}
            onChange={(e) => onChange({ ...value, max_chunk_len: numOrUndef(e.target.value) })}
          />
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <InfoLabel tip="Weight for vector (semantic) similarity in hybrid search scoring. Higher values prioritize meaning over keywords.">Vector Weight</InfoLabel>
          <Input
            type="number"
            step="0.1"
            placeholder="0.7"
            value={value.vector_weight ?? ""}
            onChange={(e) => onChange({ ...value, vector_weight: numOrUndef(e.target.value) })}
          />
        </div>
        <div className="space-y-2">
          <InfoLabel tip="Weight for text (keyword/BM25) similarity in hybrid search scoring. Higher values prioritize exact keyword matches.">Text Weight</InfoLabel>
          <Input
            type="number"
            step="0.1"
            placeholder="0.3"
            value={value.text_weight ?? ""}
            onChange={(e) => onChange({ ...value, text_weight: numOrUndef(e.target.value) })}
          />
        </div>
      </div>
    </ConfigSection>
  );
}
