import { useState } from "react"

const PRESETS = [
  { label: "15 min", value: 15 },
  { label: "30 min", value: 30 },
  { label: "1 hour", value: 60 },
  { label: "2 hours", value: 120 },
]

export function TimerControl({ value, onChange }) {
  const [custom, setCustom] = useState(value > 0 && !PRESETS.some(p => p.value === value))

  const handlePreset = (val) => {
    setCustom(false)
    onChange(val)
  }

  const handleCustomChange = (e) => {
    const v = parseInt(e.target.value, 10)
    if (!isNaN(v) && v > 0) {
      onChange(v)
    } else if (e.target.value === "") {
      onChange(0)
    }
  }

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-muted-foreground">Duration</label>
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => { setCustom(false); onChange(0) }}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-all ${
            !value ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:border-primary/50"
          }`}
        >
          No timer
        </button>
        {PRESETS.map(p => (
          <button
            key={p.value}
            onClick={() => handlePreset(p.value)}
            className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-all ${
              value === p.value && !custom
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground hover:border-primary/50"
            }`}
          >
            {p.label}
          </button>
        ))}
        <button
          onClick={() => { setCustom(true); onChange(0) }}
          className={`px-3 py-1.5 text-xs font-medium rounded-lg border transition-all ${
            custom ? "border-primary bg-primary/10 text-primary" : "border-border text-muted-foreground hover:border-primary/50"
          }`}
        >
          Custom
        </button>
      </div>
      {custom && (
        <div className="flex items-center gap-2">
          <input
            type="number"
            min="1"
            max="1440"
            value={value || ""}
            onChange={handleCustomChange}
            placeholder="Minutes"
            className="w-24 h-9 rounded-lg border border-border bg-transparent px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50"
          />
          <span className="text-xs text-muted-foreground">minutes</span>
        </div>
      )}
    </div>
  )
}
