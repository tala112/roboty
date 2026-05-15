import { useState, useEffect } from "react"
import { Search, Loader2, Monitor, CheckCircle2 } from "lucide-react"

export function AppSelector({ selected = [], onToggle, modeId }) {
  const [apps, setApps] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState("")
  const [customApp, setCustomApp] = useState("")
  const [checking, setChecking] = useState(false)
  const [customStatus, setCustomStatus] = useState(null) // null | "checking" | "found" | "not-found" | "added"

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const { modesService } = await import("../services/modes")
        const detectable = await modesService.getAllDetectableApps()
        if (!cancelled) setApps(detectable || [])
      } catch (err) {
        console.error("Failed to load apps:", err)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [])

  const filtered = search
    ? apps.filter(a =>
        (a.name || "").toLowerCase().includes(search.toLowerCase()) ||
        (a.exec || "").toLowerCase().includes(search.toLowerCase())
      )
    : apps

  const isSelected = (exec) => selected.some(s => s.app_exec === exec)

  const handleAddCustom = async () => {
    const name = customApp.trim()
    if (!name) return

    // Derive exec from name
    const exec = name.toLowerCase().replace(/[^a-z0-9._-]/g, "").replace(/\.exe$/, "")

    // Check if already in list
    const alreadyInList = apps.some(a => a.exec?.toLowerCase() === exec)
    if (alreadyInList) {
      // Just toggle it
      if (!isSelected(exec)) {
        onToggle({ name, exec })
      }
      setCustomApp("")
      return
    }

    // Check if already selected
    if (isSelected(exec)) {
      setCustomApp("")
      return
    }

    setChecking(true)
    setCustomStatus("checking")

    try {
      const { modesService } = await import("../services/modes")

      // Check if the app exists on this PC
      const onPC = await modesService.checkAppOnPC(exec)

      if (onPC) {
        setCustomStatus("found")

        // Add to mappings for persistence
        await modesService.addAllowedApp(modeId || "", name, exec, "productive", false)

        // Add to the local list
        setApps(prev => {
          const exists = prev.some(a => a.exec?.toLowerCase() === exec)
          if (exists) return prev
          return [{
            name: name,
            exec: exec,
          }, ...prev]
        })

        // Toggle it on
        onToggle({ name, exec })

        setTimeout(() => setCustomStatus("added"), 300)
        setTimeout(() => setCustomStatus(null), 1500)
      } else {
        setCustomStatus("not-found")
        // Still allow adding it — user knows their system
        // We add with force=true in case it's not detected but still exists
        await modesService.addAllowedApp(modeId || "", name, exec, "productive", true)

        setApps(prev => {
          const exists = prev.some(a => a.exec?.toLowerCase() === exec)
          if (exists) return prev
          return [{ name, exec }, ...prev]
        })
        onToggle({ name, exec })

        setTimeout(() => setCustomStatus(null), 2000)
      }
    } catch (err) {
      console.error("Failed to check app:", err)
      setCustomStatus(null)
    } finally {
      setChecking(false)
      setCustomApp("")
    }
  }

  const statusMessage = () => {
    switch (customStatus) {
      case "checking": return "Checking system..."
      case "found": return "✓ Found on PC — added to mappings"
      case "not-found": return "Not detected on PC — added anyway"
      case "added": return "✓ Added to app list"
      default: return ""
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        <Loader2 className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Loading apps from system + mappings...</span>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-muted-foreground">Allowed Apps</label>

      {/* Manual app entry */}
      <div className="flex gap-2">
        <input
          value={customApp}
          onChange={(e) => { setCustomApp(e.target.value); setCustomStatus(null) }}
          onKeyDown={(e) => e.key === "Enter" && !checking && handleAddCustom()}
          placeholder="Type app name or exec (e.g. chrome, discord)..."
          className="flex-1 h-9 rounded-lg border border-border bg-transparent px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
        <button
          onClick={handleAddCustom}
          disabled={!customApp.trim() || checking}
          className="h-9 px-3 rounded-lg bg-primary text-primary-foreground text-sm font-medium disabled:opacity-50"
        >
          {checking ? <Loader2 className="h-4 w-4 animate-spin" /> : "Add"}
        </button>
      </div>

      {/* Status message */}
      {customStatus && (
        <div className={`flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-lg ${
          customStatus === "not-found"
            ? "bg-yellow-500/10 text-yellow-500"
            : "bg-green-500/10 text-green-500"
        }`}>
          {customStatus === "checking" && <Loader2 className="h-3 w-3 animate-spin" />}
          {customStatus === "found" && <CheckCircle2 className="h-3 w-3" />}
          {statusMessage()}
        </div>
      )}

      {/* Search */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search apps..."
          className="w-full h-9 rounded-lg border border-border bg-transparent pl-9 pr-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
      </div>

      {/* App list */}
      <div className="max-h-60 overflow-y-auto rounded-lg border border-border space-y-0.5 p-1 custom-scrollbar">
        {filtered.length === 0 && (
          <p className="py-4 text-center text-xs text-muted-foreground">
            {search ? "No matching apps found" : "No apps detected"}
          </p>
        )}
        {filtered.map((app, i) => {
          const sel = isSelected(app.exec)
          const sourceLabel = app._source === "mapping" ? "mapping" : app._source === "running" ? "running" : "installed"
          return (
            <div
              key={app.exec + i}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
                sel ? "bg-green-500/10" : "hover:bg-secondary/50"
              }`}
            >
              <input
                type="checkbox"
                checked={sel}
                onChange={() => onToggle(app)}
                className="h-4 w-4 rounded border-border accent-green-500"
                title={sel ? "Allowed" : "Click to allow"}
              />
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-secondary/50">
                <Monitor className="h-4 w-4 text-muted-foreground" />
              </div>
              <span className="flex-1 truncate">{app.name}</span>
              <span className="text-[10px] text-muted-foreground bg-secondary/30 px-1.5 py-0.5 rounded font-mono">
                {app.exec}
              </span>
            </div>
          )
        })}
      </div>

      {selected.length > 0 && (
        <p className="text-xs text-muted-foreground">
          {selected.length} app{selected.length !== 1 ? "s" : ""} allowed — everything else will be blocked during focus mode
        </p>
      )}

      <style>{`
        .custom-scrollbar::-webkit-scrollbar { width: 5px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.15); border-radius: 10px; }
      `}</style>
    </div>
  )
}
