import { useState } from "react"
import { WelcomePage } from "./components/WelcomePage"
import { ChatPage } from "./components/ChatPage"
import { RunCommand } from "../wailsjs/go/main/App"

function App() {
  const [showChat, setShowChat] = useState(false)
  const [message, setMessage] = useState("")
  const [messages, setMessages] = useState([
    {
      id: "1",
      content: "Hello! I'm Roboty, your AI assistant. I'm here to help you with anything you need.",
      sender: "bot",
      timestamp: "10:30 AM",
    },
  ])

  const [isLoading, setIsLoading] = useState(false)

  const handleGoToChat = () => setShowChat(true)
  const handleClose = () => setShowChat(false)

  const handleSendMessage = async () => {
    if (!message.trim()) return
    const userMsg = {
      id: Date.now().toString(),
      content: message,
      sender: "user",
      timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
    }
    setMessages((prev) => [...prev, userMsg])
    setMessage("")
    setIsLoading(true)
    try {
      const result = await RunCommand(message)
      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: result || "Command executed successfully",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      }
      setMessages((prev) => [...prev, botMsg])
    } catch (err) {
      setMessages((prev) => [...prev, {
        id: (Date.now() + 1).toString(),
        content: "Error: " + err,
        sender: "bot",
        timestamp: "Error"
      }])
    } finally {
      setIsLoading(false)
    }
  }
  const handleNewChat = () => {
  setMessages([
    {
      id: "1",
      content: "Hello! I'm Roboty, your AI assistant.",
      sender: "bot",
      timestamp: new Date().toLocaleTimeString([], {
        hour: "2-digit",
        minute: "2-digit",
      }),
    },
  ])
  }
  if (showChat) {
    return (
      <ChatPage
        messages={messages}
        message={message}
        setMessage={setMessage}
        handleSendMessage={handleSendMessage}
        onClose={handleClose}
        handleNewChat={handleNewChat}
        isLoading={isLoading}
        RunCommand={RunCommand}
      />
    )
  }

  return <WelcomePage onGoToChat={handleGoToChat} />
}

export default App
