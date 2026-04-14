import { useState } from "react"
import { ThemeProvider } from "./components/theme-provider"
import { StarBackground } from "./components/star-background"
import { RobotAvatar } from "./components/robot-avatar"
import { RobotTitle } from "./components/robot-title"
import { Button } from "./components/ui/button"
import { ScrollArea } from "./components/ui/scroll-area"
import { Input } from "./components/ui/input"
import { ArrowRight, MessageSquare, Plus, Search, Send, User } from "lucide-react"
import { RunCommand } from "../wailsjs/go/main/App"

function App() {
  const [showChat, setShowChat] = useState(false)
  const [message, setMessage] = useState("")
  const [messages, setMessages] = useState([
    {
      id: "1",
      content:
        "Hello! I'm Roboty, your AI assistant. I'm here to help you with anything you need.",
      sender: "bot",
      timestamp: "10:30 AM",
    },
  ])

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

  if (showChat) {
    return (
      <ThemeProvider defaultTheme="dark">
        <div className="h-screen w-screen overflow-hidden bg-background relative">
          <StarBackground />

          <div className="relative z-10 flex h-full">

            {/* Sidebar */}
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

            {/* Chat Area */}
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

  return (
    <ThemeProvider defaultTheme="dark">
      <div className="h-screen w-screen bg-background relative overflow-hidden">
        <StarBackground />

        <div className="relative z-10 flex h-full w-full">
          
          {/* Left Robot */}
          <div className="w-1/2 flex items-center justify-center">
            <RobotAvatar size="xl" />
          </div>

          {/* Right Card */}
          <div className="w-1/2 flex items-center justify-center px-6">
            <div className="bg-card/50 backdrop-blur-xl border border-border/50 rounded-2xl p-8 w-full max-w-3xl flex flex-col">

              <RobotTitle />

              <ScrollArea className="mt-4 flex-1 pr-2">
                <div className="text-sm text-muted-foreground space-y-4 leading-relaxed">

                  <p className="text-foreground font-medium text-lg">
                    👋 Welcome to Your Laptop Robot
                  </p>

                  <p>
                    Turn your laptop into an intelligent assistant that can understand, monitor, and help you control your system.
                  </p>

                  <div>
                    <p className="font-semibold text-foreground mb-2">This app helps you:</p>
                    <ul className="list-disc pl-5 space-y-1">
                      <li>Talk to your computer in natural language</li>
                      <li>Monitor system performance</li>
                      <li>Analyze logs and errors</li>
                      <li>Get smart suggestions and fixes</li>
                      <li>Execute safe commands with permission</li>
                    </ul>
                  </div>

                </div>
              </ScrollArea>

              <div className="pt-4 mt-4 border-t border-border">
                <Button
                  onClick={handleGoToChat}
                  className="w-full rounded-full py-5 text-lg font-semibold"
                >
                  Start Chatting
                  <ArrowRight className="ml-2 w-5 h-5" />
                </Button>
              </div>

            </div>
          </div>

        </div>
      </div>
    </ThemeProvider>
  )
}

export default App