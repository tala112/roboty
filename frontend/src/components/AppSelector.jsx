import { useState, useEffect } from "react"
import { Search, Loader2, Monitor, Smartphone } from "lucide-react"

export function AppSelector({ selected = [], onToggle, onToggleClose }) {
  const [apps, setApps] = useState([])
  const [loading, setLoading] = useState(true)
  const [search, setSearch] = useState("")

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const { modesService } = await import("../services/modes")
        const installed = await modesService.getInstalledApps()
        if (!cancelled) setApps(installed || [])
      } catch (err) {
        console.error("Failed to load installed apps:", err)
      } finally {
        if (!cancelled) setLoading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [])

  const filtered = search
    ? apps.filter(a => a.name.toLowerCase().includes(search.toLowerCase()))
    : apps

  const isSelected = (exec) => selected.some(s => s.app_exec === exec)
  const getSelected = (exec) => selected.find(s => s.app_exec === exec)

  if (loading) {
    return (
      <div className="flex items-center justify-center py-8 text-muted-foreground">
        <Loader2 className="h-5 w-5 animate-spin mr-2" />
        <span className="text-sm">Detecting installed apps...</span>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-muted-foreground">Blocked Apps</label>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search apps..."
          className="w-full h-9 rounded-lg border border-border bg-transparent pl-9 pr-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
        />
      </div>

      <div className="max-h-60 overflow-y-auto rounded-lg border border-border space-y-0.5 p-1 custom-scrollbar">
        {filtered.length === 0 && (
          <p className="py-4 text-center text-xs text-muted-foreground">
            {search ? "No matching apps found" : "No installed apps detected"}
          </p>
        )}
        {filtered.map((app, i) => {
          const sel = isSelected(app.exec)
          const item = getSelected(app.exec)
          return (
            <div
              key={app.exec + i}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm transition-colors ${
                sel ? "bg-primary/10" : "hover:bg-secondary/50"
              }`}
            >
              <input
                type="checkbox"
                checked={sel}
                onChange={() => onToggle(app)}
                className="h-4 w-4 rounded border-border accent-primary"
              />
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-secondary/50">
                <Monitor className="h-4 w-4 text-muted-foreground" />
              </div>
              <span className="flex-1 truncate">{app.name}</span>
              {sel && (
                <label className="flex items-center gap-1.5 text-xs text-muted-foreground shrink-0">
                  <input
                    type="checkbox"
                    checked={item?.close_on_activate || false}
                    onChange={(e) => onToggleClose(app, e.target.checked)}
                    className="h-3 w-3 rounded border-border accent-primary"
                  />
                  Close
                </label>
              )}
            </div>
          )
        })}
      </div>

      <style>{`
        .custom-scrollbar::-webkit-scrollbar { width: 5px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.15); border-radius: 10px; }
      `}</style>
    </div>
  )
}
