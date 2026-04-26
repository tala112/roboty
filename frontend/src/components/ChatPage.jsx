import { useEffect, useRef, useState } from "react"
import { ThemeProvider, useTheme } from "./theme-provider"
import { StarBackground } from "./star-background"
import { RobotAvatar } from "./robot-avatar"
import { Button } from "./ui/button"
import { ScrollArea } from "./ui/scroll-area"
import { Input } from "./ui/input"
import {
  Send,
  User,
  Search,
  Plus,
  MessageSquare,
  X,
  Sun,
  Moon,
  Check,
  XCircle,
  Trash2,
  MoreVertical,
} from "lucide-react"

function ThemeToggle() {
  const { theme, setTheme } = useTheme()

  const toggleTheme = () => {
    setTheme(theme === "dark" ? "light" : "dark")
  }

  return (
    <Button variant="ghost" size="icon" onClick={toggleTheme} className="h-9 w-9">
      {theme === "dark" ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </Button>
  )
}

// Format chat time for display
function formatChatTime(dbTime) {
  if (!dbTime) return "Now"
  try {
    const date = new Date(dbTime)
    const now = new Date()
    const diff = now - date
    
    // Less than 1 minute
    if (diff < 60000) return "Just now"
    // Less than 1 hour
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    // Less than 24 hours
    if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
    // Less than 7 days
    if (diff < 604800000) return `${Math.floor(diff / 86400000)}d ago`
    // Otherwise show date
    return date.toLocaleDateString()
  } catch {
    return "Now"
  }
}

// Get chat title (first line of content)
function getChatTitle(messages) {
  if (!messages || messages.length === 0) return "Chat"
  const firstUserMsg = messages.find(m => m.sender === 'user')
  if (!firstUserMsg) return "Chat"
  const content = firstUserMsg.content
  if (content.length > 25) return content.substring(0, 25) + "..."
  return content
}

export function ChatPage({
  messages,
  message,
  setMessage,
  onClose,
  isLoading,
  handleSendMessage,
  handleNewChat,
  pendingCommand,
  isConfirmMode,
  handleConfirmExecution,
  handleCancelExecution,
  chats = [],
  currentChatId = null,
  handleSelectChat = () => {},
  handleDeleteChat = () => {},
}) {
  const messagesEndRef = useRef(null)
  const [showDeleteId, setShowDeleteId] = useState(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages, isConfirmMode])

  const currentChat = chats.find(c => c.id === currentChatId)
  const chatTitle = currentChat?.title || "Chat"

  return (
    <ThemeProvider defaultTheme="dark">
      <div className="relative h-screen w-screen overflow-hidden bg-background">
        <StarBackground />

        <div className="relative z-10 flex h-full">
          
          {/* Sidebar */}
          <div className="flex-none w-72 h-full flex-col border-r border-border bg-card/80 backdrop-blur-xl overflow-hidden">
            
            <div className="border-b border-border p-4">
              <div className="mb-4 flex items-center gap-3">
                <RobotAvatar size="sm" />
                <div>
                  <h1 className="text-lg font-bold">Roboty</h1>
                  <p className="text-xs text-muted-foreground">
                    AI Chat Assistant
                  </p>
                </div>
              </div>

              <div className="relative">
                <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder="Search chats..."
                  className="bg-secondary/50 pl-9"
                />
              </div>
            </div>

            <div className="p-3">
              <Button
                onClick={handleNewChat}
                className="w-full gap-2"
              >
                <Plus className="h-4 w-4" />
                New Chat
              </Button>
            </div>

            <ScrollArea className="flex-1 overflow-auto px-2">
              <div className="space-y-1">
                {chats.map((chat) => (
                  <div
                    key={chat.id}
                    className={`group relative w-full rounded-xl p-3 text-left hover:bg-secondary/80 cursor-pointer transition-colors ${
                      chat.id === currentChatId ? "bg-secondary" : ""
                    }`}
                    onClick={() => handleSelectChat(chat.id)}
                  >
                    <div className="flex items-center gap-3">
                      <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/20">
                        <MessageSquare className="h-5 w-5 text-accent" />
                      </div>

                      <div className="min-w-0 flex-1">
                        <div className="flex items-center justify-between">
                          <span className="truncate text-sm font-medium">
                            {chat.title || "Chat"}
                          </span>
                          <span className="text-xs text-muted-foreground">
                            {formatChatTime(chat.updated_at)}
                          </span>
                        </div>

                        <p className="truncate text-xs text-muted-foreground">
                          {chat.message_count} messages
                        </p>
                      </div>
                    </div>

                    {/* Delete button */}
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDeleteChat(chat.id)
                      }}
                      className="absolute right-2 top-1/2 -translate-y-1/2 opacity-0 group-hover:opacity-100 p-1 hover:bg-destructive/20 rounded"
                    >
                      <Trash2 className="h-4 w-4 text-destructive" />
                    </button>
                  </div>
                ))}

                {chats.length === 0 && (
                  <div className="p-4 text-center text-muted-foreground text-sm">
                    No chats yet. Start a new chat!
                  </div>
                )}
              </div>
            </ScrollArea>
          </div>

          {/* Main Chat Area */}
          <div className="flex flex-1 flex-col overflow-hidden">
            <div className="flex-none border-b border-border bg-card/50 p-4 backdrop-blur-xl">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-bold">{chatTitle}</h2>
                  <p className="text-xs text-muted-foreground">AI Chat</p>
                </div>
                <div className="flex items-center gap-2">
                  <ThemeToggle />
                  <Button variant="ghost" size="icon" onClick={onClose}>
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>

            <ScrollArea className="flex-1 overflow-auto">
              <div className="p-4">
                <div className="mx-auto max-w-4xl space-y-4">
                  {messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={`flex gap-3 ${
                        msg.sender === "user" ? "flex-row-reverse" : ""
                      }`}
                    >
                      {msg.sender === "bot" ? (
                        <img src="/icon.png" alt="face" className="w-18 h-18 object-contain shrink-0"/>
                      ) : (
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/30 shrink-0">
                          <User className="h-4 w-4 text-accent" />
                        </div>
                      )}

                      <div
                        className={`max-w-[75%] rounded-2xl px-4 py-3 ${
                          msg.sender === "bot"
                            ? "border border-border bg-card"
                            : "bg-primary text-primary-foreground"
                        }`}
                      >
                        <p className="text-sm leading-relaxed whitespace-pre-wrap">{msg.content}</p>
                        <span className="mt-2 block text-xs opacity-70">
                          {msg.timestamp}
                        </span>
                      </div>
                    </div>
                  ))}
                  <div ref={messagesEndRef} />
                </div>
              </div>
            </ScrollArea>

            {isConfirmMode && pendingCommand && (
              <div className="flex-none absolute bottom-24 right-4 z-20 mx-4 max-w-[280px] rounded-lg border border-border bg-card p-3 shadow-lg">
                <p className="text-sm mb-2">Execute this command?</p>
                <code className="text-xs bg-secondary p-1 rounded block mb-3 break-all">{pendingCommand}</code>
                <div className="flex gap-2">
                  <Button
                    onClick={handleConfirmExecution}
                    disabled={isLoading}
                    className="h-8 text-xs bg-green-600 hover:bg-green-700"
                  >
                    <Check className="h-3 w-3 mr-1" />
                    Confirm
                  </Button>
                  <Button
                    onClick={handleCancelExecution}
                    variant="outline"
                    className="h-8 text-xs"
                  >
                    <XCircle className="h-3 w-3 mr-1" />
                    Cancel
                  </Button>
                </div>
              </div>
            )}

            <div className="flex-none border-t border-border bg-card/50 p-4 backdrop-blur-xl">
              <div className="mx-auto flex max-w-4xl gap-2">
                <Input
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") {
                      handleSendMessage()
                    }
                  }}
                  placeholder="Type a message..."
                  className="h-12"
                />

                <Button
                  onClick={handleSendMessage}
                  disabled={!message.trim() || isLoading}
                  className="h-12 w-12 shrink-0"
                >
                  <Send className="h-5 w-5" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ThemeProvider>
  )
}