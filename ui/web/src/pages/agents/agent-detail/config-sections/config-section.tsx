import { Info } from "lucide-react";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";

/** Label with an (i) tooltip icon for field descriptions. */
export function InfoLabel({ children, tip }: { children: React.ReactNode; tip: string }) {
  return (
    <div className="flex items-center gap-1.5">
      <Label>{children}</Label>
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <Info className="h-3.5 w-3.5 shrink-0 cursor-help text-muted-foreground" />
          </TooltipTrigger>
          <TooltipContent side="top" className="max-w-64">
            {tip}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </div>
  );
}

interface ConfigSectionProps {
  title: string;
  description: string;
  enabled: boolean;
  onToggle: (v: boolean) => void;
  children: React.ReactNode;
}

export function ConfigSection({
  title,
  description,
  enabled,
  onToggle,
  children,
}: ConfigSectionProps) {
  return (
    <section className="space-y-3">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-sm font-medium">{title}</h3>
          <p className="text-xs text-muted-foreground">{description}</p>
        </div>
        <Switch checked={enabled} onCheckedChange={onToggle} />
      </div>
      {enabled ? (
        <div className="rounded-lg border p-4 space-y-4">{children}</div>
      ) : (
        <p className="text-xs text-muted-foreground italic">
          Using global defaults from config.json
        </p>
      )}
    </section>
  );
}

/** Parse a string input value to number or undefined. */
export function numOrUndef(v: string): number | undefined {
  const n = Number(v);
  return isNaN(n) || v === "" ? undefined : n;
}

/** Convert comma-separated string to string array, or undefined if empty. */
export function tagsToArray(s: string): string[] | undefined {
  const trimmed = s.trim();
  if (!trimmed) return undefined;
  return trimmed.split(",").map((t) => t.trim()).filter(Boolean);
}

/** Convert string array to comma-separated display string. */
export function arrayToTags(arr?: string[]): string {
  return arr?.join(", ") ?? "";
}
