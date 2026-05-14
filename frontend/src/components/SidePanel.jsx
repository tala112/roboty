import { useState, useEffect } from "react"
import { ChevronRight, ChevronLeft, FileText, Info, Terminal, ChevronDown, ChevronUp } from "lucide-react"
import { GetSessionInfo, GetSessionFiles } from "../../wailsjs/go/main/App"

function CollapsibleSection({ title, icon: Icon, children, defaultOpen = true }) {
  const [open, setOpen] = useState(defaultOpen)
  return (
    <div className="border-b border-border/50">
      <button
        onClick={() => setOpen(!open)}
        className="flex w-full items-center gap-2 px-3 py-2.5 text-sm font-medium text-muted-foreground hover:text-foreground transition-colors"
      >
        <Icon className="h-4 w-4" />
        <span className="flex-1 text-left">{title}</span>
        {open ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
      </button>
      {open && <div className="px-3 pb-3">{children}</div>}
    </div>
  )
}

export function SidePanel({ isOpen, onToggle, currentChatId }) {
  const [sessionInfo, setSessionInfo] = useState(null)
  const [files, setFiles] = useState([])
  const [toolResults, setToolResults] = useState([])

  useEffect(() => {
    if (!currentChatId) {
      setSessionInfo(null)
      setFiles([])
      return
    }
    GetSessionInfo(currentChatId).then(data => {
      try { setSessionInfo(JSON.parse(data)) } catch { setSessionInfo(null) }
    }).catch(() => setSessionInfo(null))
    GetSessionFiles(currentChatId).then(data => {
      try { setFiles(JSON.parse(data)) } catch { setFiles([]) }
    }).catch(() => setFiles([]))
  }, [currentChatId])

  return (
    <>
      {/* Toggle button */}
      <button
        onClick={onToggle}
        className="absolute right-0 top-1/2 -translate-y-1/2 z-20 flex h-10 w-5 items-center justify-center rounded-l-md border border-r-0 border-border bg-card text-muted-foreground hover:text-foreground transition-all"
        title={isOpen ? "Close panel" : "Open panel"}
      >
        {isOpen ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronLeft className="h-3.5 w-3.5" />}
      </button>

      {/* Panel */}
      <div
        className={`h-full overflow-hidden border-l border-border bg-card/80 backdrop-blur-xl transition-all duration-300 ${
          isOpen ? "w-72" : "w-0"
        }`}
      >
        <div className="h-full w-72 overflow-y-auto custom-scrollbar">
          <div className="p-3">
            <h3 className="text-sm font-semibold text-foreground mb-1">Side Panel</h3>
            <p className="text-xs text-muted-foreground mb-3">Session details and results</p>
          </div>

          <CollapsibleSection title="Tool Results" icon={Terminal}>
            {toolResults.length === 0 ? (
              <p className="text-xs text-muted-foreground py-2">No tool results yet.</p>
            ) : (
              <div className="space-y-1">
                {toolResults.map((r, i) => (
                  <div key={i} className="rounded-lg bg-secondary/30 p-2 text-xs font-mono">
                    <pre className="whitespace-pre-wrap break-all">{r}</pre>
                  </div>
                ))}
              </div>
            )}
          </CollapsibleSection>

          <CollapsibleSection title="Files" icon={FileText}>
            {!currentChatId ? (
              <p className="text-xs text-muted-foreground py-2">No active session.</p>
            ) : files.length === 0 ? (
              <p className="text-xs text-muted-foreground py-2">No files read yet.</p>
            ) : (
              <div className="space-y-1">
                {files.map((f, i) => (
                  <div key={i} className="flex items-center gap-2 rounded-lg bg-secondary/30 p-2 text-xs">
                    <FileText className="h-3 w-3 shrink-0 text-muted-foreground" />
                    <span className="truncate">{typeof f === "string" ? f : f.path || f.name || JSON.stringify(f)}</span>
                  </div>
                ))}
              </div>
            )}
          </CollapsibleSection>

          <CollapsibleSection title="Session Info" icon={Info}>
            {!currentChatId ? (
              <p className="text-xs text-muted-foreground py-2">No active session.</p>
            ) : !sessionInfo ? (
              <p className="text-xs text-muted-foreground py-2">Loading...</p>
            ) : (
              <div className="space-y-2 text-xs">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Title</span>
                  <span className="font-medium truncate ml-2">{sessionInfo.title || "Untitled"}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Messages</span>
                  <span className="font-medium">{sessionInfo.message_count || 0}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Prompt Tokens</span>
                  <span className="font-medium">{sessionInfo.prompt_tokens || 0}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Completion Tokens</span>
                  <span className="font-medium">{sessionInfo.completion_tokens || 0}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Cost</span>
                  <span className="font-medium">${(sessionInfo.cost || 0).toFixed(6)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Created</span>
                  <span className="font-medium text-xs">{sessionInfo.created_at || "-"}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Updated</span>
                  <span className="font-medium text-xs">{sessionInfo.updated_at || "-"}</span>
                </div>
              </div>
            )}
          </CollapsibleSection>
        </div>
      </div>
    </>
  )
}
