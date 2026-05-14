import { useEffect, useRef, useState } from "react"
import { useTheme } from "./theme-provider"
import { StarBackground } from "./star-background"
import { RobotAvatar } from "./robot-avatar"
import { Button } from "./ui/button"
import { Input } from "./ui/input"
import { SidePanel } from "./SidePanel"
import {
  Send,
  User,
  Search,
  Plus,
  MessageSquare,
  X,
  Sun,
  Moon,
  Trash2,
  Ellipsis,
  Settings,
} from "lucide-react"

function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === "dark" ? "light" : "dark")}
    >
      {theme === "dark" ? <Sun /> : <Moon />}
    </Button>
  )
}

// FIX: missing function
function formatChatTime(time) {
  if (!time) return ""
  try {
    return new Date(time).toLocaleTimeString([], {
      hour: "2-digit",
      minute: "2-digit",
    })
  } catch {
    return time
  }
}

export function ChatPage({
  messages,
  message,
  setMessage,
  onClose,
  isLoading,
  handleSendMessage,
  handleNewChat,
  chats = [],
  currentChatId = null,
  handleSelectChat = () => {},
  handleDeleteChat = () => {},
  onGoToSettings,
}) {
  const messagesEndRef = useRef(null)
  const [sidePanelOpen, setSidePanelOpen] = useState(false)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  const currentChat = chats.find((c) => c.id === currentChatId)
  const chatTitle = currentChat?.title || "Chat"

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-background">
      <StarBackground />

      <div className="relative z-10 flex h-full">

        {/* SIDEBAR */}
        <div className="flex w-72 flex-col border-r border-border bg-card/80 backdrop-blur-xl">

          {/* HEADER */}
          <div className="border-b border-border p-4">
            <div className="flex items-center gap-3">
              <RobotAvatar size="sm" />
              <div>
                <h1 className="font-bold">Roboty</h1>
                <p className="text-xs text-muted-foreground">AI Assistant</p>
              </div>
            </div>

            <div className="relative mt-3">
              <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted-foreground" />
              <Input className="pl-9" placeholder="Search chats..." />
            </div>
          </div>

          {/* NEW CHAT */}
          <div className="p-3">
            <Button onClick={handleNewChat} className="w-full gap-2">
              <Plus className="h-4 w-4" />
              New Chat
            </Button>
          </div>

          {/* CHAT LIST */}
          <div className="flex-1 overflow-y-auto px-2 space-y-1">

            {chats.map((chat) => (
              <div
                key={chat.id}
                onClick={() => handleSelectChat(chat.id)}
                className={`group relative flex cursor-pointer items-center gap-3 rounded-xl p-3 hover:bg-secondary ${
                  chat.id === currentChatId ? "bg-secondary" : ""
                }`}
              >

                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/20">
                  <MessageSquare className="h-5 w-5 text-accent" />
                </div>

                <div className="flex-1 min-w-0">
                  <div className="flex justify-between">
                    <span className="truncate text-sm font-medium">
                      {chat.title || "Chat"}
                    </span>
                    <span className="text-xs text-muted-foreground">
                      {formatChatTime(chat.updated_at)}
                    </span>
                  </div>
                  <p className="text-xs text-muted-foreground truncate">
                    {chat.message_count} messages
                  </p>
                </div>

                {/* DELETE */}
                <button
                  onClick={(e) => {
                    e.stopPropagation()
                    handleDeleteChat(chat.id)
                  }}
                  className="opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive/20 rounded"
                >
                  <Trash2 className="h-4 w-4 text-destructive" />
                </button>

              </div>
            ))}

            {chats.length === 0 && (
              <p className="text-center text-sm text-muted-foreground mt-4">
                No chats yet
              </p>
            )}
          </div>
        </div>

        {/* MAIN CHAT */}
        <div className="flex flex-1 flex-col">

          {/* TOP BAR */}
          <div className="flex items-center justify-between border-b border-border p-4 bg-card/50 backdrop-blur-xl">

            <div>
              <h2 className="font-bold text-lg">{chatTitle}</h2>
              <p className="text-xs text-muted-foreground">AI Chat</p>
            </div>

            <div className="flex items-center gap-2">

              <ThemeToggle />

              {onGoToSettings && (
                <Button
                  variant="ghost"
                  size="icon"
                  onClick={onGoToSettings}
                >
                  <Settings />
                </Button>
              )}

              <Button variant="ghost" size="icon" onClick={onClose}>
                <X />
              </Button>
            </div>
          </div>

          {/* MESSAGES */}
          <div className="flex-1 overflow-y-auto p-4">
            <div className="mx-auto max-w-3xl space-y-4">

              {messages.map((msg) => (
                <div
                  key={msg.id}
                  className={`flex gap-3 ${
                    msg.sender === "user" ? "flex-row-reverse" : ""
                  }`}
                >

                  {msg.sender === "bot" ? (
                    <img src="/icon.png" className="h-10 w-10" />
                  ) : (
                    <div className="h-8 w-8 flex items-center justify-center bg-accent/30 rounded-lg">
                      <User className="h-4 w-4" />
                    </div>
                  )}

                  <div
                    className={`max-w-[75%] rounded-2xl px-4 py-3 ${
                      msg.sender === "bot"
                        ? "bg-card border"
                        : "bg-primary text-white"
                    }`}
                  >
                    <p className="text-sm whitespace-pre-wrap">
                      {msg.content}
                    </p>
                    <span className="text-xs opacity-70 block mt-2">
                      {msg.timestamp}
                    </span>
                  </div>

                </div>
              ))}

              <div ref={messagesEndRef} />
            </div>
          </div>

          {/* INPUT */}
          <div className="border-t border-border p-4 bg-card/50">
            <div className="flex gap-2 max-w-3xl mx-auto">

              <Input
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
                placeholder="Type a message..."
              />

              <Button
                onClick={handleSendMessage}
                disabled={!message.trim() || isLoading}
              >
                <Send />
              </Button>

            </div>
          </div>

        </div>

        {/* SIDE PANEL */}
        <SidePanel
          isOpen={sidePanelOpen}
          onToggle={() => setSidePanelOpen(!sidePanelOpen)}
          currentChatId={currentChatId}
        />

      </div>
    </div>
  )
}