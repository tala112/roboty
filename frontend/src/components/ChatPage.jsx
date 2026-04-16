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

 /* const handleSendMessage = async () => {
    if (!message.trim()) return

    const userMsg = {
      id: Date.now().toString(),
      content: message,
      sender: "user",
      timestamp: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    }

    setMessages((prev) => [...prev, userMsg])
    setMessage("")

    try {
      const result = await RunCommand(message)

      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: result || "Command executed successfully",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        }),
      }

      setMessages((prev) => [...prev, botMsg])
    } catch (err) {
      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: "Error executing command: " + err,
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        }),
      }

      setMessages((prev) => [...prev, botMsg])
    }
  }*/

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
              {/*<Button
                onClick={() =>
                  setMessages([
                    {
                      id: "1",
                      content: "Hello! I'm Roboty, your AI assistant.",
                      sender: "bot",
                      timestamp: "10:30 AM",
                    },
                  ])
                }
                className="w-full gap-2"
              >
                <Plus className="h-4 w-4" />
                New Chat
              </Button>*/}
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
/*import { useEffect, useRef } from "react"
import { ThemeProvider } from "./theme-provider"
import { StarBackground } from "./star-background"
import { Button } from "./ui/button"
import { ScrollArea } from "./ui/scroll-area"
import { Input } from "./ui/input"
import { Send, User, X, Loader2 } from "lucide-react"

export function ChatPage({ messages, message, setMessage, handleSendMessage, onClose, isLoading }) {
  const messagesEndRef = useRef(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])



  const handleGoToChat = () => setShowChat(true)

  const handleSendMessage = async () => {
    if (!message.trim()) return

    const userMsg = {
      id: Date.now().toString(),
      content: message,
      sender: "user",
      timestamp: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    }

    setMessages((prev) => [...prev, userMsg])
    setMessage("")

    try {
      const result = await RunCommand(message)

      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: result || "Command executed successfully",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        }),
      }

      setMessages((prev) => [...prev, botMsg])
    } catch (err) {
      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: "Error executing command: " + err,
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        }),
      }

      setMessages((prev) => [...prev, botMsg])
    }
  }
    return (
      <ThemeProvider defaultTheme="dark">
        <div className="h-screen w-screen overflow-hidden bg-background relative">
          <StarBackground />

          <div className="relative z-10 flex h-full">

            {/* Sidebar * /}
            <div className="w-72 h-full bg-card/80 backdrop-blur-xl border-r border-border flex flex-col">
              <div className="p-4 border-b border-border">
                <div className="flex items-center gap-3 mb-4">
                  <RobotAvatar size="md" />
                  <div>
                    <h1 className="font-bold text-lg">Roboty</h1>
                    <p className="text-xs text-muted-foreground">AI Chat Assistant</p>
                  </div>
                </div>

                <div className="relative">
                  <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                  <Input placeholder="Search chats..." className="pl-9 bg-secondary/50" />
                </div>
              </div>

              <div className="p-3">
                <Button
                  onClick={() =>
                    setMessages([
                      {
                        id: "1",
                        content: "Hello! I'm Roboty, your AI assistant.",
                        sender: "bot",
                        timestamp: "10:30 AM",
                      },
                    ])
                  }
                  className="w-full gap-2"
                >
                  <Plus className="w-4 h-4" />
                  New Chat
                </Button>
              </div>

              <ScrollArea className="flex-1 px-2">
                <button className="w-full flex items-center gap-3 p-3 rounded-lg hover:bg-secondary bg-secondary">
                  <div className="w-10 h-10 rounded-lg bg-accent/20 flex items-center justify-center">
                    <MessageSquare className="w-5 h-5 text-accent" />
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex justify-between">
                      <span className="text-sm font-medium truncate">General Assistant</span>
                      <span className="text-xs text-muted-foreground">Now</span>
                    </div>
                    <p className="text-xs text-muted-foreground truncate">
                      Start a conversation...
                    </p>
                  </div>
                </button>
              </ScrollArea>
            </div>

            {/* Chat Area * /}
            <div className="flex-1 flex flex-col">
              <div className="p-4 border-b border-border bg-card/50 backdrop-blur-xl">
                <h2 className="font-bold text-lg">General Assistant</h2>
                <p className="text-xs text-muted-foreground">AI Chat</p>
              </div>

              <ScrollArea className="flex-1 p-4">
                <div className="space-y-6 max-w-4xl mx-auto">
                  {messages.map((msg) => (
                    <div
                      key={msg.id}
                      className={`flex gap-3 ${
                        msg.sender === "user" ? "flex-row-reverse" : ""
                      }`}
                    >
                      {msg.sender === "bot" ? (
                        <RobotAvatar size="sm" animated={false} />
                      ) : (
                        <div className="w-8 h-8 rounded-lg bg-accent/30 flex items-center justify-center">
                          <User className="w-4 h-4 text-accent" />
                        </div>
                      )}

                      <div
                        className={`max-w-[70%] rounded-2xl px-4 py-3 ${
                          msg.sender === "bot"
                            ? "bg-card border border-border"
                            : "bg-primary text-primary-foreground"
                        }`}
                      >
                        <p className="text-sm">{msg.content}</p>
                        <span className="text-xs mt-2 block opacity-70">
                          {msg.timestamp}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </ScrollArea>

              <div className="p-4 border-t border-border bg-card/50 backdrop-blur-xl">
                <div className="max-w-4xl mx-auto flex gap-2">
                  <Input
                    value={message}
                    onChange={(e) => setMessage(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter") handleSendMessage()
                    }}
                    placeholder="Type a message..."
                    className="h-12"
                  />
                  <Button
                    onClick={handleSendMessage}
                    disabled={!message.trim()}
                    className="h-12 w-12"
                  >
                    <Send className="w-5 h-5" />
                  </Button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </ThemeProvider>
    )
  }

  /*
  return (
    <ThemeProvider defaultTheme="dark">
      <div className="h-screen w-screen overflow-hidden bg-[#0a0a12] relative">
        <StarBackground />
        <div className="relative z-10 flex h-full">
          <div className="flex-1 flex flex-col">
            <div className="p-4 flex justify-between items-center border-b border-white/10">
              <div className="flex items-center gap-3">
                <img src="/icon.png" alt="icon" className="w-8 h-8" />
                <h2 className="text-white font-medium">Roboty</h2>
              </div>
              <Button variant="ghost" onClick={onClose} className="text-white hover:text-white">
                <X />
              </Button>
            </div>
            <ScrollArea className="flex-1 p-6">
              <div className="space-y-4 max-w-2xl mx-auto">
                {messages.map((msg) => (
                  <div
                    key={msg.id}
                    className={`flex ${msg.sender === "user" ? "justify-end" : "justify-start"}`}
                  >
                    <div
                      className={`flex gap-3 max-w-[80%] ${
                        msg.sender === "user" ? "flex-row-reverse" : "flex-row"
                      }`}
                    >
                      <div
                        className={`w-8 h-8 rounded-full flex items-center justify-center shrink-0 ${
                          msg.sender === "bot"
                            ? "bg-white/10"
                            : "bg-gray-600"
                        }`}
                      >
                        {msg.sender === "bot" ? (
                          <img src="/icon.png" className="w-5 h-5" alt="bot" />
                        ) : (
                          <User className="w-4 h-4 text-white" />
                        )}
                      </div>
                      <div>
                        <div
                          className={`rounded-2xl px-4 py-2 ${
                            msg.sender === "bot"
                              ? "bg-white/10 text-white"
                              : "bg-gray-600 text-white"
                          }`}
                        >
                          <p className="whitespace-pre-wrap">{msg.content}</p>
                        </div>
                        <p className="text-xs text-gray-500 mt-1 px-1">
                          {msg.timestamp}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
                {isLoading && (
                  <div className="flex justify-start">
                    <div className="flex gap-3">
                      <div className="w-8 h-8 rounded-full bg-white/10 flex items-center justify-center">
                        <img src="/icon.png" className="w-5 h-5" alt="bot" />
                      </div>
                      <div className="bg-white/10 rounded-2xl px-4 py-3">
                        <Loader2 className="w-5 h-5 text-white animate-spin" />
                      </div>
                    </div>
                  </div>
                )}
                <div ref={messagesEndRef} />
              </div>
            </ScrollArea>
            <div className="p-4 border-t border-white/10">
              <div className="flex gap-3 max-w-2xl mx-auto">
                <Input
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
                  placeholder="Ask Roboty anything..."
                  className="flex-1 bg-white/10 border-white/20 text-white placeholder:text-gray-400 rounded-full"
                />
                <Button
                  onClick={handleSendMessage}
                  disabled={!message.trim() || isLoading}
                  className="rounded-full px-6 bg-white text-black hover:bg-gray-200"
                >
                  <Send className="w-4 h-4" />
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ThemeProvider>
  )*/

