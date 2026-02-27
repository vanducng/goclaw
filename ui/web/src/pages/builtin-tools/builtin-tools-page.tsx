import { useState, useEffect } from "react";
import { Package, RefreshCw, Settings } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { PageHeader } from "@/components/shared/page-header";
import { EmptyState } from "@/components/shared/empty-state";
import { SearchInput } from "@/components/shared/search-input";
import { Pagination } from "@/components/shared/pagination";
import { TableSkeleton } from "@/components/shared/loading-skeleton";
import { useBuiltinTools, type BuiltinToolData } from "./hooks/use-builtin-tools";
import { BuiltinToolSettingsDialog } from "./builtin-tool-settings-dialog";
import { useMinLoading } from "@/hooks/use-min-loading";
import { useDeferredLoading } from "@/hooks/use-deferred-loading";
import { usePagination } from "@/hooks/use-pagination";

export function BuiltinToolsPage() {
  const { tools, loading, refresh, updateTool } = useBuiltinTools();
  const spinning = useMinLoading(loading);
  const showSkeleton = useDeferredLoading(loading && tools.length === 0);
  const [search, setSearch] = useState("");
  const [settingsTool, setSettingsTool] = useState<BuiltinToolData | null>(null);

  const filtered = tools.filter(
    (t) =>
      t.name.toLowerCase().includes(search.toLowerCase()) ||
      t.display_name.toLowerCase().includes(search.toLowerCase()) ||
      t.description.toLowerCase().includes(search.toLowerCase()) ||
      t.category.toLowerCase().includes(search.toLowerCase()),
  );

  const { pageItems, pagination, setPage, setPageSize, resetPage } = usePagination(filtered, { defaultPageSize: 50 });

  useEffect(() => {
    resetPage();
  }, [search, resetPage]);

  const handleToggle = async (tool: BuiltinToolData) => {
    await updateTool(tool.name, { enabled: !tool.enabled });
  };

  const handleSaveSettings = async (name: string, settings: Record<string, unknown>) => {
    await updateTool(name, { settings });
  };

  const hasSettings = (tool: BuiltinToolData) =>
    tool.settings && Object.keys(tool.settings).length > 0;

  const categories = [...new Set(tools.map((t) => t.category))].sort();

  return (
    <div className="p-6">
      <PageHeader
        title="Built-in Tools"
        description="Manage system built-in tools. Enable/disable tools or configure their settings globally."
        actions={
          <Button
            variant="outline"
            size="sm"
            onClick={refresh}
            disabled={spinning}
            className="gap-1"
          >
            <RefreshCw className={"h-3.5 w-3.5" + (spinning ? " animate-spin" : "")} />
            Refresh
          </Button>
        }
      />

      <div className="mt-4 flex items-center gap-3">
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder="Search by name, description, or category..."
          className="max-w-sm"
        />
        <div className="text-sm text-muted-foreground">
          {filtered.length} tool{filtered.length !== 1 ? "s" : ""}
          {categories.length > 0 && ` across ${categories.length} categories`}
        </div>
      </div>

      <div className="mt-4">
        {showSkeleton ? (
          <TableSkeleton rows={8} />
        ) : filtered.length === 0 ? (
          <EmptyState
            icon={Package}
            title={search ? "No matching tools" : "No built-in tools"}
            description={
              search ? "Try a different search term." : "No built-in tools have been seeded yet."
            }
          />
        ) : (
          <div className="rounded-md border">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b bg-muted/50">
                  <th className="px-4 py-3 text-left font-medium">Name</th>
                  <th className="px-4 py-3 text-left font-medium">Description</th>
                  <th className="px-4 py-3 text-left font-medium">Category</th>
                  <th className="px-4 py-3 text-center font-medium">Enabled</th>
                  <th className="px-4 py-3 text-right font-medium">Actions</th>
                </tr>
              </thead>
              <tbody>
                {pageItems.map((tool) => (
                  <tr
                    key={tool.name}
                    className="border-b last:border-0 hover:bg-muted/30"
                  >
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Package className="h-4 w-4 shrink-0 text-muted-foreground" />
                        <div>
                          <span className="font-medium">{tool.display_name}</span>
                          <span className="ml-2 text-xs text-muted-foreground">{tool.name}</span>
                        </div>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-muted-foreground">
                      {tool.description || "-"}
                    </td>
                    <td className="px-4 py-3">
                      <Badge variant="outline">{tool.category}</Badge>
                    </td>
                    <td className="px-4 py-3 text-center">
                      <Switch
                        checked={tool.enabled}
                        onCheckedChange={() => handleToggle(tool)}
                      />
                    </td>
                    <td className="px-4 py-3 text-right">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSettingsTool(tool)}
                        className="gap-1"
                      >
                        <Settings className="h-3.5 w-3.5" />
                        {hasSettings(tool) ? "Edit" : "Settings"}
                      </Button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            <Pagination
              page={pagination.page}
              pageSize={pagination.pageSize}
              total={pagination.total}
              totalPages={pagination.totalPages}
              onPageChange={setPage}
              onPageSizeChange={setPageSize}
            />
          </div>
        )}
      </div>

      <BuiltinToolSettingsDialog
        tool={settingsTool}
        open={settingsTool !== null}
        onOpenChange={(open) => {
          if (!open) setSettingsTool(null);
        }}
        onSave={handleSaveSettings}
      />
    </div>
  );
}
