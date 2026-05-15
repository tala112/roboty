import { useState } from "react"

const PRESETS = [
  { label: "15 min", value: 15 },
  { label: "30 min", value: 30 },
  { label: "1 hour", value: 60 },
  { label: "2 hours", value: 120 },
]

export function TimerControl({ value, onChange }) {
  const [inputVal, setInputVal] = useState(value > 0 ? String(value) : "")

  const handlePreset = (val) => {
    setInputVal(String(val))
    onChange(val)
  }

  const handleInputChange = (e) => {
    const raw = e.target.value
    // Allow only digits
    if (!/^\d*$/.test(raw)) return
    setInputVal(raw)
    if (raw === "") {
      onChange(0)
    } else {
      const v = parseInt(raw, 10)
      if (v > 0) {
        onChange(v)
      }
    }
  }

  const handleBlur = () => {
    if (inputVal === "" || inputVal === "0") {
      setInputVal("")
      onChange(0)
    }
  }

  return (
    <div className="space-y-3">
      <label className="block text-sm font-medium text-muted-foreground">Duration</label>
      <div className="flex flex-wrap gap-2">
        <button
          onClick={() => { setInputVal(""); onChange(0) }}
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
              value === p.value
                ? "border-primary bg-primary/10 text-primary"
                : "border-border text-muted-foreground hover:border-primary/50"
            }`}
          >
            {p.label}
          </button>
        ))}
      </div>
      <div className="flex items-center gap-2">
        <input
          type="text"
          inputMode="numeric"
          value={inputVal}
          onChange={handleInputChange}
          onBlur={handleBlur}
          placeholder="Minutes (any duration)"
          className="w-36 h-9 rounded-lg border border-border bg-transparent px-3 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/50 placeholder:text-muted-foreground/50"
        />
        <span className="text-xs text-muted-foreground">minutes</span>
        {value > 0 && (
          <span className="text-xs text-muted-foreground/60">
            ({value >= 60 ? `${Math.floor(value / 60)}h ${value % 60}m` : `${value} min`})
          </span>
        )}
      </div>
    </div>
  )
}
