import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import type { ToolPolicyConfig } from "@/types/agent";
import { ConfigSection, InfoLabel, arrayToTags, tagsToArray } from "./config-section";

interface ToolPolicySectionProps {
  enabled: boolean;
  value: ToolPolicyConfig;
  onToggle: (v: boolean) => void;
  onChange: (v: ToolPolicyConfig) => void;
}

export function ToolPolicySection({ enabled, value, onToggle, onChange }: ToolPolicySectionProps) {
  return (
    <ConfigSection
      title="Tool Policy"
      description="Control which tools this agent can use"
      enabled={enabled}
      onToggle={onToggle}
    >
      <div className="space-y-2">
        <InfoLabel tip="Base tool profile. 'default' includes common tools, 'strict' limits to safe tools only, 'permissive' allows all tools.">Profile</InfoLabel>
        <Select
          value={value.profile ?? ""}
          onValueChange={(v) => onChange({ ...value, profile: v || undefined })}
        >
          <SelectTrigger><SelectValue placeholder="default" /></SelectTrigger>
          <SelectContent>
            <SelectItem value="default">default</SelectItem>
            <SelectItem value="strict">strict</SelectItem>
            <SelectItem value="permissive">permissive</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <InfoLabel tip="Comma-separated allowlist. Only these tools will be available (overrides profile). Leave empty to use profile defaults.">Allow</InfoLabel>
        <Input
          placeholder="tool1, tool2, ..."
          value={arrayToTags(value.allow)}
          onChange={(e) => onChange({ ...value, allow: tagsToArray(e.target.value) })}
        />
      </div>
      <div className="space-y-2">
        <InfoLabel tip="Comma-separated denylist. These tools will be blocked even if allowed by the profile.">Deny</InfoLabel>
        <Input
          placeholder="tool1, tool2, ..."
          value={arrayToTags(value.deny)}
          onChange={(e) => onChange({ ...value, deny: tagsToArray(e.target.value) })}
        />
      </div>
      <div className="space-y-2">
        <InfoLabel tip="Additional tools added on top of the profile defaults. Useful for enabling optional tools like web_fetch without overriding the whole profile.">Also Allow</InfoLabel>
        <Input
          placeholder="web_fetch, web_search, ..."
          value={arrayToTags(value.alsoAllow)}
          onChange={(e) => onChange({ ...value, alsoAllow: tagsToArray(e.target.value) })}
        />
      </div>
    </ConfigSection>
  );
}
