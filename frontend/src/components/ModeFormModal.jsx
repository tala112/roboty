import { useState, useEffect } from "react"
import { Modal } from "./ui/modal"
import { TimerControl } from "./TimerControl"
import { AppSelector } from "./AppSelector"
import { Input } from "./ui/input"
import { Button } from "./ui/button"
import { Save, Plus, X } from "lucide-react"

const ICON_OPTIONS = [
  { value: "shield", label: "Shield" },
  { value: "timer", label: "Timer" },
  { value: "bell", label: "Bell" },
  { value: "eye", label: "Eye" },
  { value: "zap", label: "Zap" },
  { value: "moon", label: "Moon" },
]

const COLOR_OPTIONS = [
  "#6366f1", "#8b5cf6", "#a855f7", "#ec4899",
  "#ef4444", "#f97316", "#eab308", "#22c55e",
  "#14b8a6", "#06b6d4", "#3b82f6", "#64748b",
]

function DefaultIcon() {
  return (
    <svg className="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
    </svg>
  )
}

export function ModeFormModal({ open, onClose, mode, onSave }) {
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [duration, setDuration] = useState(0)
  const [muteNotif, setMuteNotif] = useState(false)
  const [icon, setIcon] = useState("shield")
  const [color, setColor] = useState("#6366f1")
  const [apps, setApps] = useState([])
  const [allowedUrls, setAllowedUrls] = useState([])
  const [urlInput, setUrlInput] = useState("")
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (open) {
      if (mode) {
        setName(mode.name || "")
        setDescription(mode.description || "")
        setDuration(mode.duration_minutes || 0)
        setMuteNotif(mode.mute_notifications || false)
        setIcon(mode.icon || "shield")
        setColor(mode.color || "#6366f1")
        setApps((mode.apps || []).map(a => ({
          app_name: a.app_name,
          app_exec: a.app_exec,
          close_on_activate: a.close_on_activate || false,
          is_allowed: a.is_allowed !== false,
        })))
        setAllowedUrls(mode.allowed_urls || [])
      } else {
        setName("")
        setDescription("")
        setDuration(0)
        setMuteNotif(false)
        setIcon("shield")
        setColor("#6366f1")
        setApps([])
        setAllowedUrls([])
      }
    }
  }, [open, mode])

  const handleToggleApp = (app) => {
    setApps(prev => {
      const exists = prev.find(a => a.app_exec === app.exec)
      if (exists) {
        return prev.filter(a => a.app_exec !== app.exec)
      }
      return [{
        app_name: app.name,
        app_exec: app.exec,
        close_on_activate: false,
        is_allowed: true,
      }, ...prev]
    })
  }

  const handleToggleClose = (app, close) => {
    setApps(prev => prev.map(a =>
      a.app_exec === app.exec
        ? { ...a, close_on_activate: close }
        : a
    ))
  }

  const handleAddUrl = () => {
    const url = urlInput.trim().toLowerCase()
    if (!url || allowedUrls.includes(url)) return
    setAllowedUrls(prev => [...prev, url])
    setUrlInput("")
  }

  const handleRemoveUrl = (url) => {
    setAllowedUrls(prev => prev.filter(u => u !== url))
  }

  const handleSubmit = async () => {
    if (!name.trim()) return
    setSaving(true)
    try {
      await onSave({
        id: mode?.id,
        name: name.trim(),
        description: description.trim(),
        durationMinutes: duration,
        muteNotifications: muteNotif,
        enabled: mode?.enabled || false,
        icon,
        color,
        apps,
        allowedUrls,
      })
      onClose()
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal open={open} onClose={onClose} title={mode ? "Edit Mode" : "Create Mode"} size="lg">
      <div className="space-y-5">
        <div className="space-y-2">
          <label className="block text-sm font-medium text-muted-foreground">Name</label>
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="e.g. Deep Work"
            className="w-full"
          />
        </div>

        <div className="space-y-2">
          <label className="block text-sm font-medium text-muted-foreground">Description</label>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="What is this mode for?"
            rows={2}
            className="w-full rounded-lg border border-border bg-transparent px-3 py-2 text-sm text-foreground resize-none focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
        </div>

        <TimerControl value={duration} onChange={setDuration} />

        <div className="flex items-center gap-2">
          <input
            type="checkbox"
            checked={muteNotif}
            onChange={(e) => setMuteNotif(e.target.checked)}
            className="h-4 w-4 rounded border-border accent-primary"
          />
          <label className="text-sm text-muted-foreground cursor-pointer">Mute notifications when active</label>
        </div>

        <div className="space-y-2">
          <label className="block text-sm font-medium text-muted-foreground">Icon</label>
          <div className="flex gap-2 flex-wrap">
            {ICON_OPTIONS.map(opt => (
              <button
                key={opt.value}
                onClick={() => setIcon(opt.value)}
                className={`flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-xs font-medium transition-all ${
                  icon === opt.value
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border text-muted-foreground hover:border-primary/50"
                }`}
              >
                <DefaultIcon />
                {opt.label}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-2">
          <label className="block text-sm font-medium text-muted-foreground">Color</label>
          <div className="flex gap-2 flex-wrap">
            {COLOR_OPTIONS.map(c => (
              <button
                key={c}
                onClick={() => setColor(c)}
                className={`h-7 w-7 rounded-lg border-2 transition-all ${
                  color === c ? "border-foreground scale-110" : "border-transparent"
                }`}
                style={{ backgroundColor: c }}
              />
            ))}
          </div>
        </div>

        {/* Allowed URLs (whitelist) */}
        <div className="space-y-2">
          <label className="block text-sm font-medium text-muted-foreground">Allowed Websites</label>
          <p className="text-xs text-muted-foreground">All sites are blocked except those listed below</p>
          <div className="flex gap-2">
            <input
              value={urlInput}
              onChange={(e) => setUrlInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleAddUrl()}
              placeholder="e.g. github.com"
              className="flex-1 h-9 rounded-lg border border-border bg-transparent px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
            <button
              onClick={handleAddUrl}
              disabled={!urlInput.trim()}
              className="h-9 px-3 rounded-lg bg-primary text-primary-foreground text-sm font-medium disabled:opacity-50"
            >
              <Plus className="h-4 w-4" />
            </button>
          </div>
          {allowedUrls.length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-1">
              {allowedUrls.map(url => (
                <span
                  key={url}
                  className="inline-flex items-center gap-1 rounded-full bg-green-500/10 text-green-500 px-2.5 py-0.5 text-xs"
                >
                  {url}
                  <button onClick={() => handleRemoveUrl(url)} className="hover:text-green-300">
                    <X className="h-3 w-3" />
                  </button>
                </span>
              ))}
            </div>
          )}
        </div>

        <AppSelector
          selected={apps}
          onToggle={handleToggleApp}
          onToggleClose={handleToggleClose}
          modeId={mode?.id || ''}
        />

        <div className="flex justify-end gap-3 pt-2">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit} disabled={!name.trim() || saving} className="gap-2">
            <Save className="h-4 w-4" />
            {saving ? "Saving..." : (mode ? "Update" : "Create")}
          </Button>
        </div>
      </div>
    </Modal>
  )
}
