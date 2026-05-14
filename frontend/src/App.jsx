import { useState, useEffect } from "react"
import { ThemeProvider } from "./components/theme-provider"
import { WelcomePage } from "./components/WelcomePage"
import { ChatPage } from "./components/ChatPage"
import { SettingsPage } from "./components/SettingsPage"
import { PreviewCommand, ExecuteCommand, InitDatabase, GetChats, GetActiveChat, GetChatMessages, SaveMessage, CreateChat, DeleteChat } from "../wailsjs/go/main/App"
import { PermissionModal } from "./components/PermissionModal"

export default function App() {
  const [showChat, setShowChat] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [isDbReady, setIsDbReady] = useState(false)
  
  const [chats, setChats] = useState([])
  const [currentChatId, setCurrentChatId] = useState(null)
  
  const [message, setMessage] = useState("")
  const [msgListFrontend, setMsgListFrontend] = useState([])

  const [isLoading, setIsLoading] = useState(false)
  const [pendingCommand, setPendingCommand] = useState(null)
  const [isConfirmMode, setIsConfirmMode] = useState(false)
  const [pendingCommandDangerous, setPendingCommandDangerous] = useState(false)

  // Extract text from JSON content
  const extractText = (content) => {
    try {
      const parsed = JSON.parse(content)
      return parsed.text || content
    } catch {
      return content
    }
  }

  // Format time from DB format
  const formatTime = (dbTime) => {
    if (!dbTime) return ""
    try {
      const date = new Date(dbTime)
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    } catch {
      return dbTime
    }
  }

  // Convert DB messages to frontend format
  const convertMessages = (messages) => {
    if (!messages || messages.length === 0) return []
    return messages.map(m => ({
      id: m.id,
      content: extractText(m.content),
      sender: m.role === 'user' ? 'user' : 'bot',
      timestamp: formatTime(m.created_at),
    }))
  }

  // Initialize database and load chats on startup
  useEffect(() => {
    const init = async () => {
      try {
        await InitDatabase()
        setIsDbReady(true)
        
        // Load all chats for sidebar display
        const allChats = await GetChats()
        if (allChats && allChats.length > 0) {
          setChats(allChats)
        }
        
        // Always start with NEW empty chat (don't load previous messages)
        setCurrentChatId(null)
        setMsgListFrontend([{
          id: "welcome-" + Date.now(),
          content: "Hello! I'm Roboty, your AI assistant.",
          sender: "bot",
          timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        }])
      } catch (err) {
        console.error("Init error:", err)
      }
    }
    init()
  }, [])

  // Handle chat selection - save current chat first, then load new chat
  const handleSelectChat = async (chatId) => {
    // Note: Messages are already saved in real-time via saveToDb on each message
    // So we just need to refresh the chats list and load the new chat
    setCurrentChatId(chatId)
    setMsgListFrontend([])
    setIsConfirmMode(false)
    setPendingCommand(null)
    
    try {
      // Refresh chats list from DB to get latest updates (including titles)
      const allChats = await GetChats()
      setChats(allChats)
      
      // Load messages for selected chat
      const messages = await GetChatMessages(chatId)
      setMsgListFrontend(convertMessages(messages))
    } catch (err) {
      console.error("Error loading messages:", err)
    }
  }

  // Handle new chat - save current messages, refresh list, create new chat
  const handleNewChat = async () => {
    try {
      // Note: Current messages are already saved in real-time via saveToDb
      // Just refresh the chats list to ensure we have latest state
      const allChatsBefore = await GetChats()
      setChats(allChatsBefore)
      
      // Create new chat
      const newChat = await CreateChat("")
      
      // Refresh chats list from DB (will include new chat)
      const allChats = await GetChats()
      setChats(allChats)
      
      setCurrentChatId(newChat.id)
      setMsgListFrontend([{
        id: "welcome-" + Date.now(),
        content: "Hello! I'm Roboty, your AI assistant.",
        sender: "bot",
        timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
      }])
      setIsConfirmMode(false)
      setPendingCommand(null)
    } catch (err) {
      console.error("Error creating chat:", err)
    }
  }

  // Handle delete chat
  const handleDeleteChat = async (chatId) => {
    try {
      await DeleteChat(chatId)
      
      // Refresh chats list from DB
      const allChats = await GetChats()
      setChats(allChats)
      
      if (allChats.length > 0) {
        handleSelectChat(allChats[0].id)
      } else {
        // Create new chat if all deleted
        const newChat = await CreateChat("")
        
        // Refresh chats list from DB
        const newChatsList = await GetChats()
        setChats(newChatsList)
        
        setCurrentChatId(newChat.id)
        setMsgListFrontend([{
          id: "welcome-" + Date.now(),
          content: "Hello! I'm Roboty, your AI assistant.",
          sender: "bot",
          timestamp: new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" }),
        }])
      }
      setCurrentChatId(allChats.length > 0 ? allChats[0].id : null)
    } catch (err) {
      console.error("Error deleting chat:", err)
    }
  }

  // Save message to database
  const saveToDb = async (role, content) => {
    if (!currentChatId) return
    try {
      await SaveMessage(currentChatId, role, JSON.stringify({ text: content }))
      
      // After saving, refresh chats list to get updated titles (auto-updated by backend)
      const allChats = await GetChats()
      setChats(allChats)
    } catch (err) {
      console.error("Error saving message:", err)
    }
  }

  const handleSendMessage = async () => {
    if (!message.trim()) return
    const lastMsg = msgListFrontend[msgListFrontend.length - 1]
    if (lastMsg && lastMsg.isConfirm) return
    
    const userCmd = message.trim()
    const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    const userMsgId = Date.now().toString()
    
    setMessage("")
    
    // If no current chat (first message), create a new chat first
    let chatId = currentChatId
    if (!chatId) {
      try {
        const newChat = await CreateChat("")
        const allChats = await GetChats()
        setChats(allChats)
        chatId = newChat.id
        setCurrentChatId(chatId)
      } catch (err) {
        console.error("Error creating chat:", err)
        return
      }
    }
    
    // Save user message to DB (async, don't await for display)
    saveToDb('user', userCmd)
    
    // Add user message to UI immediately
    const userMsgObj = {
      id: userMsgId,
      content: userCmd,
      sender: "user",
      timestamp: now,
    }
    setMsgListFrontend([...msgListFrontend, userMsgObj])
    
    // Always ask for confirmation before executing
    try {
      const preview = await PreviewCommand(userCmd)

      setPendingCommand(userCmd)
      setPendingCommandDangerous(preview.is_dangerous)
      setIsConfirmMode(true)
    } catch (err) {
      console.error("Preview error:", err)
    }
  }

  const handleConfirmExecution = async () => {
    if (!pendingCommand) return
    setIsConfirmMode(false)
    setPendingCommandDangerous(false)
    
    try {
      // Execute the command
      const result = await ExecuteCommand(pendingCommand, true)
      const responseText = result || "Command executed successfully"
      
      // Save bot response to DB
      await saveToDb('assistant', responseText)
      
      // Add bot response to UI
      const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
      const botMsg = {
        id: (Date.now() + 1).toString(),
        content: responseText,
        sender: "bot",
        timestamp: now,
      }
      setMsgListFrontend(prev => [...prev, botMsg])
    } catch (err) {
      console.error("Error:", err)
    } finally {
      setPendingCommand(null)
    }
  }

  const handleCancelExecution = async () => {
    const now = new Date().toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
    const cancelMsg = { 
      id: Date.now().toString(), 
      content: "Execution cancelled by user.", 
      sender: "bot", 
      timestamp: now,
    }
    setMsgListFrontend(prev => [...prev, cancelMsg])
    
    // Save bot response to DB
    try {
      await saveToDb('assistant', "Execution cancelled by user.")
    } catch (err) {
      console.error("Error saving cancel message:", err)
    }
    
    setPendingCommand(null)
    setPendingCommandDangerous(false)
    setIsConfirmMode(false)
  }

  const onGoToChat = () => {
    setShowChat(true)
  }

  const onGoToSettings = () => {
    setShowSettings(true)
  }

  const onBackFromSettings = () => {
    setShowSettings(false)
  }

  return (
    <ThemeProvider defaultTheme="dark">
      <PermissionModal
        isOpen={isConfirmMode && !!pendingCommand}
        command={pendingCommand}
        isDangerous={pendingCommandDangerous}
        onAllow={handleConfirmExecution}
        onDeny={handleCancelExecution}
      />
      {showSettings ? (
        <SettingsPage onBack={onBackFromSettings} />
      ) : showChat ? (
        <ChatPage
          messages={msgListFrontend}
          message={message}
          setMessage={setMessage}
          handleSendMessage={handleSendMessage}
          onClose={() => setShowChat(false)}
          handleNewChat={handleNewChat}
          isLoading={isLoading}
          chats={chats}
          currentChatId={currentChatId}
          handleSelectChat={handleSelectChat}
          handleDeleteChat={handleDeleteChat}
          onGoToSettings={onGoToSettings}
        />
      ) : (
        <WelcomePage onGoToChat={onGoToChat} />
      )}
    </ThemeProvider>
  )
}