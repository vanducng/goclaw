import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { SubagentsConfig } from "@/types/agent";
import { ConfigSection, InfoLabel, numOrUndef } from "./config-section";

interface SubagentsSectionProps {
  enabled: boolean;
  value: SubagentsConfig;
  onToggle: (v: boolean) => void;
  onChange: (v: SubagentsConfig) => void;
}

export function SubagentsSection({ enabled, value, onToggle, onChange }: SubagentsSectionProps) {
  return (
    <ConfigSection
      title="Subagents"
      description="Controls sub-agent spawning limits and behavior"
      enabled={enabled}
      onToggle={onToggle}
    >
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <InfoLabel tip="Maximum number of sub-agents that can run simultaneously for this agent.">Max Concurrent</InfoLabel>
          <Input
            type="number"
            placeholder="8"
            value={value.maxConcurrent ?? ""}
            onChange={(e) => onChange({ ...value, maxConcurrent: numOrUndef(e.target.value) })}
          />
        </div>
        <div className="space-y-2">
          <InfoLabel tip="How many levels deep sub-agents can spawn other sub-agents. Depth 1 means only the parent can spawn.">Max Spawn Depth</InfoLabel>
          <Select
            value={String(value.maxSpawnDepth ?? "")}
            onValueChange={(v) => onChange({ ...value, maxSpawnDepth: Number(v) })}
          >
            <SelectTrigger><SelectValue placeholder="1" /></SelectTrigger>
            <SelectContent>
              {[1, 2, 3, 4, 5].map((n) => (
                <SelectItem key={n} value={String(n)}>{n}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <InfoLabel tip="Maximum number of sub-agents a single parent agent can spawn in one session.">Max Children Per Agent</InfoLabel>
          <Input
            type="number"
            placeholder="5"
            value={value.maxChildrenPerAgent ?? ""}
            onChange={(e) => onChange({ ...value, maxChildrenPerAgent: numOrUndef(e.target.value) })}
          />
        </div>
        <div className="space-y-2">
          <InfoLabel tip="Idle time in minutes before a sub-agent session is automatically archived and cleaned up.">Archive After (minutes)</InfoLabel>
          <Input
            type="number"
            placeholder="60"
            value={value.archiveAfterMinutes ?? ""}
            onChange={(e) => onChange({ ...value, archiveAfterMinutes: numOrUndef(e.target.value) })}
          />
        </div>
      </div>
      <div className="space-y-2">
        <InfoLabel tip="LLM model for sub-agents. Leave empty to inherit the parent agent's model.">Model Override</InfoLabel>
        <Input
          placeholder="(inherit from agent)"
          value={value.model ?? ""}
          onChange={(e) => onChange({ ...value, model: e.target.value || undefined })}
        />
      </div>
    </ConfigSection>
  );
}
