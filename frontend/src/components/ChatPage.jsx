import { useEffect, useRef, useState } from "react";
import { useTheme } from "./theme-provider";
import { StarBackground } from "./star-background";
import { RobotAvatar } from "./robot-avatar";
import { Button } from "./ui/button";
import { Input } from "./ui/input";
import { SidePanel } from "./SidePanel";
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
} from "lucide-react";

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
  const messagesEndRef = useRef(null);
  const [sidePanelOpen, setSidePanelOpen] = useState(false);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  const currentChat = chats.find((c) => c.id === currentChatId);
  const chatTitle = currentChat?.title || "Chat";

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-background">
      <StarBackground />

      <div className="relative z-10 flex h-full">
        {/* Sidebar */}
        <div className="flex w-72 flex-col overflow-hidden border-r border-border bg-card/80 backdrop-blur-xl">
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
            <Button onClick={handleNewChat} className="w-full gap-2 transition-all">
              <Plus className="h-4 w-4" />
              New Chat
            </Button>
          </div>

          <div className="flex-1 overflow-y-auto px-2 pb-2 custom-scrollbar">
            <div className="space-y-1">
              {chats.map((chat) => (
                <div
                  key={chat.id}
                  className={`group relative w-full rounded-xl p-3 text-left transition-all duration-200 hover:bg-secondary/80 cursor-pointer ${
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

                  <div className="absolute right-3 top-5/8 ">
                    <div className="relative group/dots">
                      <button
                        onClick={(e) => e.stopPropagation()}
                        className="opacity-0 transition-all duration-200 group-hover:opacity-100 rounded-md p-1 hover:bg-secondary/50"
                        aria-label="More options"
                      >
                        <Ellipsis className="h-4 w-4 text-muted-foreground" />
                      </button>

                      <div className="absolute right-0 top-full mt-0 hidden group-hover/dots:block hover:block z-[10]">
                        <div className="w-40 rounded-lg border border-border bg-card shadow-lg overflow-hidden">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteChat(chat.id);
                            }}
                            className="w-full flex items-center gap-2 px-3 py-2 text-sm text-destructive hover:bg-destructive/10 transition-colors"
                          >
                            <Trash2 className="h-4 w-4" />
                            Delete Chat
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}

              {chats.length === 0 && (
                <div className="p-4 text-center text-sm text-muted-foreground">
                  No chats yet. Start a new chat!
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Main Chat + Side Panel container */}
        <div className="relative flex flex-1 overflow-hidden">
          <div className="flex flex-1 flex-col">
            <div className="flex-none border-b border-border bg-card/50 p-4 backdrop-blur-xl">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-lg font-bold">{chatTitle}</h2>
                  <p className="text-xs text-muted-foreground">AI Chat</p>
                </div>
                <div className="flex items-center gap-2">
                  <ThemeToggle />
                  {onGoToSettings && (
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={onGoToSettings}
                      className="transition-all hover:scale-105"
                      title="Settings"
                    >
                      <Settings className="h-4 w-4" />
                    </Button>
                  )}
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={onClose}
                    className="transition-all hover:scale-105"
                  >
                    <X className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </div>

            <div className="flex-1 overflow-y-auto custom-scrollbar">
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
                        <img
                          src="/icon.png"
                          alt="face"
                          className="h-12 w-12 shrink-0 object-contain"
                        />
                      ) : (
                        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-accent/30">
                          <User className="h-4 w-4 text-accent" />
                        </div>
                      )}

                      <div
                        className={`max-w-[75%] rounded-2xl px-4 py-3 transition-all ${
                          msg.sender === "bot"
                            ? "border border-border bg-card"
                            : "bg-primary text-primary-foreground"
                        }`}
                      >
                        <p className="whitespace-pre-wrap text-sm leading-relaxed">
                          {msg.content}
                        </p>
                        <span className="mt-2 block text-xs opacity-70">
                          {msg.timestamp}
                        </span>
                      </div>
                    </div>
                  ))}
                  <div ref={messagesEndRef} />
                </div>
              </div>
            </div>

            <div className="flex-none border-t border-border bg-card/50 p-4 backdrop-blur-xl">
              <div className="mx-auto flex max-w-4xl gap-2">
                <Input
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === "Enter") handleSendMessage();
                  }}
                  placeholder="Type a message..."
                  className="h-12 transition-all"
                />
                <Button
                  onClick={handleSendMessage}
                  disabled={!message.trim() || isLoading}
                  className="h-12 w-12 shrink-0 transition-all hover:scale-105"
                >
                  <Send className="h-5 w-5" />
                </Button>
              </div>
            </div>
          </div>
          <SidePanel
            isOpen={sidePanelOpen}
            onToggle={() => setSidePanelOpen(!sidePanelOpen)}
            currentChatId={currentChatId}
          />
        </div>
      </div>

      <style>{`
        .custom-scrollbar::-webkit-scrollbar { width: 6px; }
        .custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
        .custom-scrollbar::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.2); border-radius: 10px; }
        .custom-scrollbar::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.3); }
      `}</style>
    </div>
  );
}
