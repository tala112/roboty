import { useState, useEffect } from "react"
import { ThemeProvider } from "./components/theme-provider"
import { WelcomePage } from "./components/WelcomePage"
import { ChatPage } from "./components/ChatPage"
import { RunCommand, PreviewCommand, ExecuteCommand } from "../wailsjs/go/main/App"

export default function App() {
  const [showChat, setShowChat] = useState(false)
  
  const [message, setMessage] = useState("")
  const [msgList, setMsgList] = useState([
    {
      id: "1",
      content: "Hello! I'm Roboty, your AI assistant. I'm here to help you with anything you need.",
      sender: "bot",
      timestamp: "10:30 AM",
    },
  ])

  const [isLoading, setIsLoading] = useState(false)
  const [pendingCommand, setPendingCommand] = useState(null)
  const [isConfirmMode, setIsConfirmMode] = useState(false)

  const handleSendMessage = async () => {
    if (!message.trim()) return
    const lastMsg = msgList[msgList.length - 1]
    if (lastMsg && lastMsg.isConfirm) return
    
    const userCmd = message.trim()
    setMessage("")
    setIsLoading(true)
    
    try {
      const preview = await PreviewCommand(userCmd)
      const userMsg = {
        id: Date.now().toString(),
        content: userCmd,
        sender: "user",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      }
      setPendingCommand(userCmd)
      setIsConfirmMode(true)
      const previewMsg = {
        id: (Date.now() + 1).toString(),
        content: preview.is_dangerous ? "⛔ Command blocked for security" : "⏳ Executing command...",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        isConfirm: true,
        isBlocked: preview.is_dangerous,
      }
      setMsgList([...msgList, userMsg, previewMsg])
      setIsLoading(false)
    } catch (err) {
      setIsLoading(false)
    }
  }

  const handleConfirmExecution = async () => {
    if (!pendingCommand) return
    setIsConfirmMode(false)
    setIsLoading(true)
    try {
      const result = await ExecuteCommand(pendingCommand, true)
      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: result || "Command executed successfully",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      }
      setMsgList([...msgList, botMsg])
    } catch (err) {
      setMsgList([...msgList, { id: Date.now().toString(), content: "Error: " + err, sender: "bot", timestamp: "Error" }])
    } finally {
      setPendingCommand(null)
      setIsLoading(false)
    }
  }

  const handleCancelExecution = () => {
    setMsgList([...msgList, { id: Date.now().toString(), content: "Execution cancelled by user.", sender: "bot", timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) }])
    setPendingCommand(null)
    setIsConfirmMode(false)
  }

  const handleNewChat = () => {
    setMsgList([{ id: "1", content: "Hello! I'm Roboty, your AI assistant.", sender: "bot", timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }) }])
  }

  const onGoToChat = () => {
    setShowChat(true)
  }

  console.log("Rendering App, showChat:", showChat)

  return (
    <ThemeProvider defaultTheme="dark">
      {showChat ? (
        <ChatPage
          messages={msgList}
          message={message}
          setMessage={setMessage}
          handleSendMessage={handleSendMessage}
          onClose={() => setShowChat(false)}
          handleNewChat={handleNewChat}
          isLoading={isLoading}
          RunCommand={RunCommand}
          pendingCommand={pendingCommand}
          isConfirmMode={isConfirmMode}
          handleConfirmExecution={handleConfirmExecution}
          handleCancelExecution={handleCancelExecution}
        />
      ) : (
        <WelcomePage onGoToChat={onGoToChat} />
      )}
    </ThemeProvider>
  )
}