import { useState, useEffect, useCallback } from "react";
import { useHttp } from "@/hooks/use-ws";

export interface BuiltinToolData {
  name: string;
  display_name: string;
  description: string;
  category: string;
  enabled: boolean;
  settings: Record<string, unknown>;
  requires: string[];
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export function useBuiltinTools() {
  const http = useHttp();
  const [tools, setTools] = useState<BuiltinToolData[]>([]);
  const [loading, setLoading] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await http.get<{ tools: BuiltinToolData[] }>("/v1/tools/builtin");
      setTools(res.tools ?? []);
    } catch {
      // ignore
    } finally {
      setLoading(false);
    }
  }, [http]);

  useEffect(() => {
    load();
  }, [load]);

  const updateTool = useCallback(
    async (name: string, data: { enabled?: boolean; settings?: Record<string, unknown> }) => {
      await http.put(`/v1/tools/builtin/${name}`, data);
      await load();
    },
    [http, load],
  );

  return { tools, loading, refresh: load, updateTool };
}
