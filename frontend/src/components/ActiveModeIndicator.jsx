import { useEffect, useState } from "react"
import { Timer, Shield, X } from "lucide-react"

export function ActiveModeIndicator({ session, mode, onStop }) {
  const [remaining, setRemaining] = useState("")

  useEffect(() => {
    if (!session?.ends_at) {
      setRemaining("")
      return
    }
    function tick() {
      const end = new Date(session.ends_at).getTime()
      const now = Date.now()
      const diff = end - now
      if (diff <= 0) {
        setRemaining("Ending...")
        return
      }
      const m = Math.floor(diff / 60000)
      const s = Math.floor((diff % 60000) / 1000)
      setRemaining(`${m}:${s.toString().padStart(2, "0")}`)
    }
    tick()
    const interval = setInterval(tick, 1000)
    return () => clearInterval(interval)
  }, [session])

  if (!session) return null

  return (
    <div className="flex items-center gap-3 rounded-xl border border-green-500/30 bg-green-500/10 px-4 py-2.5 backdrop-blur-sm">
      <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-green-500/20">
        <Shield className="h-4 w-4 text-green-500" />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-sm font-medium text-foreground truncate">
          {mode?.name || "Focus Mode"} Active
        </p>
        {remaining && (
          <p className="text-xs text-muted-foreground flex items-center gap-1">
            <Timer className="h-3 w-3" />
            {remaining} remaining
          </p>
        )}
      </div>
      <button
        onClick={onStop}
        className="flex items-center gap-1.5 rounded-lg bg-destructive/10 px-3 py-1.5 text-xs font-medium text-destructive hover:bg-destructive/20 transition-colors"
      >
        <X className="h-3.5 w-3.5" />
        Stop
      </button>
    </div>
  )
}
