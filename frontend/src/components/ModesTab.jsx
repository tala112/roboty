import { useState, useEffect, useCallback } from "react"
import { Plus, Power, PowerOff, Pencil, Trash2, Shield, Timer, Bell, Eye, Zap, Moon, Loader2 } from "lucide-react"
import { Button } from "./ui/button"
import { ActiveModeIndicator } from "./ActiveModeIndicator"
import { ModeFormModal } from "./ModeFormModal"

const ICON_MAP = {
  shield: Shield,
  timer: Timer,
  bell: Bell,
  eye: Eye,
  zap: Zap,
  moon: Moon,
}

function FallbackIcon() {
  return <Shield className="h-4 w-4" />
}

function ModeIcon({ icon }) {
  const Icon = ICON_MAP[icon] || FallbackIcon
  return <Icon className="h-4 w-4" />
}

export function ModesTab() {
  const [modes, setModes] = useState([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [editMode, setEditMode] = useState(null)
  const [activeSession, setActiveSession] = useState(null)
  const [activeMode, setActiveMode] = useState(null)

  const loadModes = useCallback(async () => {
    try {
      const { modesService } = await import("../services/modes")
      const [allModes, session] = await Promise.all([
        modesService.list(),
        modesService.getActiveSession(),
      ])
      setModes(allModes)
      setActiveSession(session)
      if (session) {
        const m = allModes.find(m => m.id === session.mode_id)
        setActiveMode(m || null)
      }
    } catch (err) {
      console.error("Failed to load modes:", err)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    loadModes()
  }, [loadModes])

  const handleCreate = () => {
    setEditMode(null)
    setShowForm(true)
  }

  const handleEdit = (mode) => {
    setEditMode(mode)
    setShowForm(true)
  }

  const handleSave = async (data) => {
    const { modesService } = await import("../services/modes")
    if (data.id) {
      await modesService.update(data)
    } else {
      await modesService.create(data)
    }
    await loadModes()
  }

  const handleDelete = async (id) => {
    const { modesService } = await import("../services/modes")
    await modesService.delete(id)
    await loadModes()
  }

  const handleToggle = async (id, enabled) => {
    const { modesService } = await import("../services/modes")
    if (enabled) {
      await modesService.activate(id)
      setActiveSession(null)
    } else if (activeSession?.mode_id === id) {
      await modesService.deactivate(activeSession.id)
    }
    await loadModes()
  }

  const handleActivate = async (mode) => {
    const { modesService } = await import("../services/modes")
    if (activeSession) {
      await modesService.deactivate(activeSession.id)
    }
    const session = await modesService.activate(mode.id)
    if (session) {
      setActiveSession(session)
      setActiveMode(mode)
    }
    await loadModes()
  }

  const handleStop = async () => {
    if (!activeSession) return
    const { modesService } = await import("../services/modes")
    await modesService.deactivate(activeSession.id)
    setActiveSession(null)
    setActiveMode(null)
    await loadModes()
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-16">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-xl font-bold">Focus Modes</h3>
          <p className="text-sm text-muted-foreground">Create and manage focus modes to block distractions</p>
        </div>
        <Button onClick={handleCreate} className="gap-2">
          <Plus className="h-4 w-4" />
          New Mode
        </Button>
      </div>

      {activeSession && (
        <ActiveModeIndicator
          session={activeSession}
          mode={activeMode}
          onStop={handleStop}
        />
      )}

      {modes.length === 0 && !loading && (
        <div className="rounded-2xl border border-border/50 bg-card/50 p-8 text-center">
          <div className="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-xl bg-secondary/50">
            <Shield className="h-6 w-6 text-muted-foreground" />
          </div>
          <p className="text-sm font-medium mb-1">No focus modes yet</p>
          <p className="text-xs text-muted-foreground mb-4">
            Create your first mode to block distractions and stay focused
          </p>
          <Button onClick={handleCreate} variant="outline" size="sm" className="gap-2">
            <Plus className="h-4 w-4" />
            Create Mode
          </Button>
        </div>
      )}

      <div className="space-y-3">
        {modes.map(mode => {
          const isActive = activeSession?.mode_id === mode.id
          return (
            <div
              key={mode.id}
              className={`group relative rounded-2xl border p-4 transition-all ${
                isActive
                  ? "border-green-500/40 bg-green-500/5 shadow-sm shadow-green-500/10"
                  : "border-border/50 bg-card/90 hover:border-border"
              }`}
            >
              <div className="flex items-start gap-4">
                <div
                  className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl"
                  style={{ backgroundColor: mode.color + "20", color: mode.color }}
                >
                  <ModeIcon icon={mode.icon} />
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <h4 className="text-sm font-semibold">{mode.name}</h4>
                    {isActive && (
                      <span className="rounded-full bg-green-500/20 px-2 py-0.5 text-[10px] font-medium text-green-500">
                        Active
                      </span>
                    )}
                  </div>
                  {mode.description && (
                    <p className="text-xs text-muted-foreground mt-0.5 truncate">{mode.description}</p>
                  )}
                  <div className="flex items-center gap-3 mt-2">
                    {mode.duration_minutes > 0 && (
                      <span className="flex items-center gap-1 text-[11px] text-muted-foreground">
                        <Timer className="h-3 w-3" />
                        {mode.duration_minutes >= 60
                          ? `${Math.floor(mode.duration_minutes / 60)}h ${mode.duration_minutes % 60}m`
                          : `${mode.duration_minutes} min`}
                      </span>
                    )}
                    {mode.mute_notifications && (
                      <span className="flex items-center gap-1 text-[11px] text-muted-foreground">
                        <Bell className="h-3 w-3" />
                        Muted
                      </span>
                    )}
                    <span className="text-[11px] text-muted-foreground">
                      {mode.apps?.length || 0} app{(mode.apps?.length || 0) !== 1 ? "s" : ""}
                    </span>
                  </div>
                </div>

                <div className="flex items-center gap-1 shrink-0">
                  {isActive ? (
                    <button
                      onClick={handleStop}
                      className="rounded-lg p-2 text-destructive hover:bg-destructive/10 transition-colors"
                      title="Deactivate"
                    >
                      <PowerOff className="h-4 w-4" />
                    </button>
                  ) : (
                    <button
                      onClick={() => handleActivate(mode)}
                      className="rounded-lg p-2 text-muted-foreground hover:text-green-500 hover:bg-green-500/10 transition-colors"
                      title="Activate"
                    >
                      <Power className="h-4 w-4" />
                    </button>
                  )}
                  <button
                    onClick={() => handleEdit(mode)}
                    className="rounded-lg p-2 text-muted-foreground hover:text-foreground hover:bg-secondary transition-colors"
                    title="Edit"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleDelete(mode.id)}
                    className="rounded-lg p-2 text-muted-foreground hover:text-destructive hover:bg-destructive/10 transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>
            </div>
          )
        })}
      </div>

      <ModeFormModal
        open={showForm}
        onClose={() => { setShowForm(false); setEditMode(null) }}
        mode={editMode}
        onSave={handleSave}
      />
    </div>
  )
}
