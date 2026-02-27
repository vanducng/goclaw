import { useState, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import type { BuiltinToolData } from "./hooks/use-builtin-tools";

interface Props {
  tool: BuiltinToolData | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSave: (name: string, settings: Record<string, unknown>) => Promise<void>;
}

export function BuiltinToolSettingsDialog({ tool, open, onOpenChange, onSave }: Props) {
  const [json, setJson] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    if (tool) {
      setJson(JSON.stringify(tool.settings ?? {}, null, 2));
      setError("");
    }
  }, [tool]);

  const handleSave = async () => {
    if (!tool) return;
    try {
      const parsed = JSON.parse(json);
      setSaving(true);
      setError("");
      await onSave(tool.name, parsed);
      onOpenChange(false);
    } catch (e) {
      setError(e instanceof SyntaxError ? "Invalid JSON" : String(e));
    } finally {
      setSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Settings: {tool?.display_name ?? tool?.name}</DialogTitle>
        </DialogHeader>
        <div className="space-y-3">
          <p className="text-sm text-muted-foreground">
            Edit the JSON settings for this tool. For media tools, you can set{" "}
            <code className="text-xs">provider</code> and <code className="text-xs">model</code>.
          </p>
          <Textarea
            value={json}
            onChange={(e) => setJson(e.target.value)}
            rows={10}
            className="font-mono text-sm"
            placeholder='{"provider": "gemini", "model": "gemini-2.0-flash"}'
          />
          {error && <p className="text-sm text-destructive">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)}>
            Cancel
          </Button>
          <Button onClick={handleSave} disabled={saving}>
            {saving ? "Saving..." : "Save"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
