import { useEffect, useRef } from "react"
import { ThemeProvider } from "./theme-provider"
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
} from "lucide-react"

export function ChatPage({
  messages,
  setMessages,
  message,
  setMessage,
  onClose,
  isLoading,
  handleSendMessage,
  handleNewChat,
  RunCommand,
}) {
  const messagesEndRef = useRef(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }, [messages])

  return (
    <ThemeProvider defaultTheme="dark">
      <div className="relative h-screen w-screen overflow-hidden bg-background">
        <StarBackground />

        <div className="relative z-10 flex h-full">
          
          {/* Sidebar */}
          <div className="flex h-full w-72 flex-col border-r border-border bg-card/80 backdrop-blur-xl">
            
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

            <ScrollArea className="flex-1 px-2">
              <button className="w-full rounded-xl bg-secondary p-3 text-left hover:bg-secondary/80">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-accent/20">
                    <MessageSquare className="h-5 w-5 text-accent" />
                  </div>

                  <div className="min-w-0 flex-1">
                    <div className="flex items-center justify-between">
                      <span className="truncate text-sm font-medium">
                        General Assistant
                      </span>
                      <span className="text-xs text-muted-foreground">
                        Now
                      </span>
                    </div>

                    <p className="truncate text-xs text-muted-foreground">
                      Start a conversation...
                    </p>
                  </div>
                </div>
              </button>
            </ScrollArea>
          </div>

          {/* Main Chat Area */}
          <div className="flex min-w-0 flex-1 flex-col">
            
            <div className="border-b border-border bg-card/50 p-4 backdrop-blur-xl">
              <h2 className="text-lg font-bold">General Assistant</h2>
              <p className="text-xs text-muted-foreground">AI Chat</p>
            </div>

            <ScrollArea className="flex-1 p-4">
              <div className="mx-auto max-w-4xl space-y-6">
                {messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={`flex gap-3 ${
                      msg.sender === "user" ? "flex-row-reverse" : ""
                    }`}
                  >
                    {msg.sender === "bot" ? (
                      <img src="/icon.png" alt="face" className="w-18 h-18 object-contain"/>
                    ) : (
                      <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-accent/30">
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
                      <p className="text-sm leading-relaxed">{msg.content}</p>
                      <span className="mt-2 block text-xs opacity-70">
                        {msg.timestamp}
                      </span>
                    </div>
                  </div>
                ))}

                <div ref={messagesEndRef} />
              </div>
            </ScrollArea>

            <div className="border-t border-border bg-card/50 p-4 backdrop-blur-xl">
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