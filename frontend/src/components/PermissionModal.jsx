import { Button } from "./ui/button"
import { Shield, AlertTriangle, Check, X } from "lucide-react"

export function PermissionModal({
  isOpen,
  command,
  isDangerous,
  onAllow,
  onDeny,
}) {
  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="mx-4 w-full max-w-md rounded-2xl border border-border bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95">
        <div className="flex items-center gap-3 mb-4">
          <div className={`flex h-10 w-10 items-center justify-center rounded-full ${
            isDangerous ? "bg-destructive/10" : "bg-primary/10"
          }`}>
            {isDangerous
              ? <AlertTriangle className="h-5 w-5 text-destructive" />
              : <Shield className="h-5 w-5 text-primary" />
            }
          </div>
          <div>
            <h3 className="text-lg font-bold">Command Permission</h3>
            <p className="text-sm text-muted-foreground">
              {isDangerous
                ? "This command appears unsafe"
                : "Execute this command?"
              }
            </p>
          </div>
        </div>

        {isDangerous && (
          <div className="mb-3 rounded-lg bg-destructive/10 p-3 border border-destructive/20">
            <p className="text-xs text-destructive font-medium">
              This command matches dangerous patterns. Only allow if you trust it completely.
            </p>
          </div>
        )}

        {command && (
          <div className="mb-4 rounded-lg bg-secondary/30 p-3">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-sm font-medium">Command</span>
            </div>
            <code className="block mt-1 rounded bg-secondary p-2 text-xs font-mono break-all">
              {command}
            </code>
          </div>
        )}

        <div className="flex gap-2 justify-end">
          <Button variant="outline" onClick={onDeny} className="gap-2">
            <X className="h-4 w-4" />
            Deny
          </Button>
          <Button onClick={onAllow} className={`gap-2 ${
            isDangerous ? "bg-red-600 hover:bg-red-700" : "bg-green-600 hover:bg-green-700"
          }`}>
            <Check className="h-4 w-4" />
            Allow
          </Button>
        </div>
      </div>
    </div>
  )
}
