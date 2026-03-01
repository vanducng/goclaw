import { useState, useCallback, useEffect } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { useWsCall } from "@/hooks/use-ws-call";
import { useWsEvent } from "@/hooks/use-ws-event";

interface ZaloPersonalQRDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  instanceId: string;
  instanceName: string;
  onSuccess: () => void;
}

export function ZaloPersonalQRDialog({
  open,
  onOpenChange,
  instanceId,
  instanceName,
  onSuccess,
}: ZaloPersonalQRDialogProps) {
  const [qrPng, setQrPng] = useState<string | null>(null);
  const [status, setStatus] = useState<"idle" | "waiting" | "done" | "error">("idle");
  const [errorMsg, setErrorMsg] = useState("");
  const { call: startQR, loading } = useWsCall("zalo.personal.qr.start");

  const handleStart = useCallback(async () => {
    setStatus("waiting");
    setQrPng(null);
    setErrorMsg("");
    try {
      await startQR({ instance_id: instanceId });
    } catch (err) {
      setStatus("error");
      setErrorMsg(err instanceof Error ? err.message : "Failed to start QR session");
    }
  }, [startQR, instanceId]);

  // Auto-start when dialog opens
  useEffect(() => {
    if (open && status === "idle") handleStart();
  }, [open]); // eslint-disable-line react-hooks/exhaustive-deps

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setStatus("idle");
      setQrPng(null);
      setErrorMsg("");
    }
  }, [open]);

  useWsEvent(
    "zalo.personal.qr.code",
    useCallback(
      (payload: unknown) => {
        const p = payload as { instance_id: string; png_b64: string };
        if (p.instance_id !== instanceId) return;
        setQrPng(p.png_b64);
      },
      [instanceId],
    ),
  );

  useWsEvent(
    "zalo.personal.qr.done",
    useCallback(
      (payload: unknown) => {
        const p = payload as { instance_id: string; success: boolean; error?: string };
        if (p.instance_id !== instanceId) return;
        if (p.success) {
          setStatus("done");
          onSuccess();
          setTimeout(() => onOpenChange(false), 1500);
        } else {
          setStatus("error");
          setErrorMsg(p.error ?? "QR login failed");
        }
      },
      [instanceId, onSuccess, onOpenChange],
    ),
  );

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!loading) onOpenChange(v); }}>
      <DialogContent className="max-w-sm">
        <DialogHeader>
          <DialogTitle>Login with QR — {instanceName}</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col items-center gap-4 py-4">
          {status === "done" && (
            <p className="text-sm text-green-600 font-medium">Login successful! Channel starting...</p>
          )}
          {status === "error" && (
            <p className="text-sm text-destructive">{errorMsg}</p>
          )}
          {qrPng && status === "waiting" && (
            <img
              src={`data:image/png;base64,${qrPng}`}
              alt="Zalo QR Code"
              className="w-48 h-48 border rounded"
            />
          )}
          {status === "waiting" && !qrPng && (
            <p className="text-sm text-muted-foreground">Generating QR code...</p>
          )}
          {status === "waiting" && qrPng && (
            <p className="text-xs text-muted-foreground">Scan with your Zalo app (expires in ~100s)</p>
          )}
        </div>

        <div className="flex justify-end gap-2">
          <Button variant="outline" onClick={() => onOpenChange(false)}>Close</Button>
          {status === "error" && (
            <Button onClick={handleStart} disabled={loading}>Retry</Button>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
