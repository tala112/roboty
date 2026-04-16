import { ThemeProvider } from "./theme-provider"
import { StarBackground } from "./star-background"
import { RobotAvatar } from "./robot-avatar"
import { ArrowRight } from "lucide-react"

export function WelcomePage({ onGoToChat }) {
  return (
    <ThemeProvider defaultTheme="dark">
      <div className="h-screen w-screen relative overflow-hidden bg-[#050508] font-sans">
        <div className="absolute inset-0 z-0">
          <StarBackground />
        </div>

        <div className="absolute top-8 left-8 z-20 flex flex-col items-start">
          <img src="/titlewithstar.png" alt="" className="w-24 sm:w-28 md:w-30 lg:w-80 h-auto object-contain"/>
        </div>

        <div className="absolute top-100 right-280 z-30">
          <RobotAvatar size="xl" animated={true} />
        </div>

          <div className="absolute inset-0 flex items-center justify-center md:justify-end md:pr-20 p-6 z-20">
          <div 
            className="w-full max-w-5xl rounded-[40px] p-8 md:p-12 xl:py-20 shadow-2xl transition-all"
            style={{ 
              backgroundColor: 'rgba(13, 13, 18, 0.85)', 
              backdropFilter: 'blur(12px)',
              border: '1px solid rgba(255, 255, 255, 0.1)' 
            }}
          >
            <h1 className="text-2xl md:text-3xl text-white font-light mb-6 leading-relaxed">
              Welcome to Your <span className="font-semibold">Laptop Robot</span>.
            </h1>
            
            <p className="text-gray-300 text-lg mb-8 leading-snug">
              Turn your laptop into an intelligent assistant that can understand, monitor, and help you control your system.
            </p>

            <div className="space-y-4">
              <h2 className="text-gray-200 font-medium text-lg">This app helps you:</h2>
              <ul className="space-y-3">
                {[
                  "Talk to your computer in natural language",
                  "Monitor system performance",
                  "Analyze logs and errors",
                  "Get smart suggestions and fixes",
                  "Execute safe commands with permission"
                ].map((item, i) => (
                  <li key={i} className="flex items-start text-gray-400 text-base">
                    <span className="mr-3 mt-1.5 w-1.5 h-1.5 rounded-full bg-gray-500 shrink-0" />
                    {item}
                  </li>
                ))}
              </ul>
            </div>

            <div className="mt-10 flex justify-end">
              <button
                onClick={onGoToChat}
                className="flex items-center gap-2 bg-white text-black px-6 py-2.5 rounded-full font-bold hover:bg-gray-200 transition-all active:scale-95 shadow-lg"
              >
                go to chat
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        <img src="/icon.png" className="absolute bottom-4 right-4 w-6 opacity-20" alt="" />
      </div>
    </ThemeProvider>
  )
}
