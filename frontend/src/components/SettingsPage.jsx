import { useState, useEffect, Component } from "react"
import { StarBackground } from "./star-background"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { useTheme } from "./theme-provider"
import {
  Settings, ArrowLeft, Plus, Trash2, Save, Globe,
  Server, Terminal, Shield, Code, Sun, Moon, Cpu,
  FileText, Info, Check, X ,Circle,Star,
} from "lucide-react"
import { ModesTab } from "./ModesTab"
import { GetChats } from "../../wailsjs/go/main/App"

class ErrorBoundary extends Component {
  constructor(props) {
    super(props)
    this.state = { hasError: false, errorMessage: "" }
  }
  static getDerivedStateFromError(error) {
    return { hasError: true, errorMessage: error?.message || "Unknown error" }
  }
  componentDidCatch(error, info) {
    console.error("Tab error:", error, info)
  }
  render() {
    if (this.state.hasError) {
      return (
        <div className="p-8 text-center text-muted-foreground">
          <p className="text-lg font-medium mb-2">Something went wrong</p>
          <p className="text-xs mb-1 font-mono bg-destructive/10 text-destructive p-2 rounded">{this.state.errorMessage}</p>
          <p className="text-sm mb-4">There was an error loading this section.</p>
          <Button onClick={() => this.setState({ hasError: false })}>Retry</Button>
        </div>
      )
    }
    return this.props.children
  }
}

function TabButton({ active, onClick, children }) {
  return (
    <button
      onClick={onClick}
      className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium rounded-lg transition-all ${
        active
          ? "bg-primary text-primary-foreground shadow-sm"
          : "text-muted-foreground hover:text-foreground hover:bg-secondary/50"
      }`}
    >
      {children}
    </button>
  )
}

const CircleStarIcon = () => (
  <div className="relative w-6 h-6">
    <Circle className="absolute" size={20} />
    <Star className="absolute top-1 left-1" size={12} />
  </div>
);

function InputField({ label, value, onChange, placeholder, type = "text", className = "" }) {
  return (
    <div className={className}>
      <label className="block text-sm font-medium mb-1 text-muted-foreground">{label}</label>
      <Input
        type={type}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full"
      />
    </div>
  )
}



function Toggle({ enabled, onChange, label }) {
  return (
    <label className="flex items-center gap-3 cursor-pointer">
      <div
        onClick={onChange}
        className={`relative w-10 h-5 rounded-full transition-colors ${
          enabled ? "bg-primary" : "bg-secondary"
        }`}
      >
        <div
          className={`absolute top-0.5 left-0.5 w-4 h-4 rounded-full bg-white transition-transform ${
            enabled ? "translate-x-5" : "translate-x-0"
          }`}
        />
      </div>
      <span className="text-sm">{label}</span>
    </label>
  )
}

export function SettingsPage({ onBack }) {
  const { theme, setTheme } = useTheme()
  const [activeTab, setActiveTab] = useState("providers")

  const tabs = [
    { id:"mode", label:"Modes", icon: CircleStarIcon },
    { id: "general", label: "General", icon: Settings },
  ]

  return (
    <div className="h-screen w-screen overflow-hidden bg-background">
      <StarBackground />
      <div className="flex h-full">
        <div className="w-56 flex flex-col border-r border-border bg-card/50 p-4">
          <div className="flex items-center gap-2 mb-6">
            <Settings className="h-5 w-5" />
            <h2 className="font-bold text-lg">Settings</h2>
          </div>
          <nav className="flex flex-col gap-1 flex-1">
            {tabs.map((tab) => (
              <TabButton
                key={tab.id}
                active={activeTab === tab.id}
                onClick={() => setActiveTab(tab.id)}
              >
                <tab.icon className="h-4 w-4" />
                {tab.label}
              </TabButton>
            ))}
          </nav>
          <Button
            variant="outline"
            onClick={onBack}
            className="mt-4 gap-2"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to Chat
          </Button>
        </div>

        <div className="flex flex-col flex-1 overflow-hidden">
          <div className="flex items-center justify-between border-b border-border px-6 py-3 bg-card/30">
            <h3 className="text-sm font-medium text-muted-foreground">
              {tabs.find(t => t.id === activeTab)?.label || "Settings"}
            </h3>
            <Button variant="ghost" size="icon" onClick={onBack} className="h-8 w-8">
              <X className="h-4 w-4" />
            </Button>
          </div>
          <div className="flex-1 overflow-y-auto p-6">
            <div className="max-w-3xl mx-auto">
              <ErrorBoundary key={activeTab}>
                {activeTab === "mode" && <ModesTab /> }
                {activeTab === "general" && <GeneralTab theme={theme} setTheme={setTheme} onBack={onBack} />}
              </ErrorBoundary>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function GeneralTab({ theme, setTheme, onBack }) {
  const [agentType, setAgentType] = useState("coder")
  const [version, setVersion] = useState("")
  const [initStatus, setInitStatus] = useState("")

  useEffect(() => {
    setVersion("0.1.0")
  }, [])

  const handleAgentChange = (val) => {
    setAgentType(val)
  }

  const handleInit = () => {
    setInitStatus("Init triggered (not implemented)")
    setTimeout(() => setInitStatus(""), 3000)
  }

  return (
    <div className="space-y-8">
      <div>
        <h3 className="text-xl font-bold">General Settings</h3>
        <p className="text-sm text-muted-foreground">App-wide preferences</p>
      </div>

      <div className="rounded-2xl border border-border/50 bg-card/90 backdrop-blur-xl shadow-xl p-5 space-y-6">
        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium">Theme</p>
            <p className="text-xs text-muted-foreground">Dark or light mode</p>
          </div>
          <Button variant="outline" onClick={() => setTheme(theme === "dark" ? "light" : "dark")} className="gap-2">
            {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
            {theme === "dark" ? "Light" : "Dark"}
          </Button>
        </div>

        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium">Agent Type</p>
            <p className="text-xs text-muted-foreground">Choose between coder and task agent</p>
          </div>
          <div className="flex gap-2">
            <Button variant={agentType === "coder" ? "default" : "outline"} size="sm" onClick={() => handleAgentChange("coder")}>Coder</Button>
            <Button variant={agentType === "task" ? "default" : "outline"} size="sm" onClick={() => handleAgentChange("task")}>Task</Button>
          </div>
        </div>

        <div className="flex items-center justify-between">
          <div>
            <p className="font-medium">Version</p>
            <p className="text-xs text-muted-foreground">App version {version}</p>
          </div>
        </div>
      </div>

      <div className="rounded-2xl border border-border/50 bg-card/90 backdrop-blur-xl shadow-xl p-5 space-y-4">
        <h4 className="font-medium">Actions</h4>
        <div className="flex flex-wrap gap-3">
          <Button variant="outline" onClick={handleInit} className="gap-2">
            <FileText className="h-4 w-4" /> Init Project
          </Button>
          <Button variant="outline" onClick={() => console.log("View Logs clicked")} className="gap-2">
            <Info className="h-4 w-4" /> View Logs
          </Button>
        </div>
        {initStatus && <p className="text-sm text-green-500">{initStatus}</p>}
      </div>
    </div>
  )
}
