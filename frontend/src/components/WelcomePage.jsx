/*
import { StarBackground } from "./star-background"
import { RobotAvatar } from "./robot-avatar"
import { ArrowRight } from "lucide-react"

export function WelcomePage({ onGoToChat }) {
  return (
      <div className="h-screen w-screen relative overflow-hidden bg-[#050508] font-sans">
        <div className="absolute inset-0 z-0">
          <StarBackground />
        </div>

          <div className="absolute inset-0 flex items-center justify-center p-6 z-20">
          <div 
            className="w-full max-w-4xl rounded-[40px] p-8 md:p-12 xl:py-20 shadow-2xl transition-all"
            style={{ 
              backgroundColor: 'rgba(13, 13, 18, 0.85)', 
              backdropFilter: 'blur(12px)',
              border: '1px solid rgba(255, 255, 255, 0.1)' 
            }}
          >
              <div className="absolute top-8 left-8 z-20 flex flex-col items-start">
              <img src="/titlewithstar.png" alt="" className="w-24 sm:w-28 md:w-30 lg:w-80 h-auto object-contain"/>
              </div>

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

         <div className="absolute left-180 top-0 z-30">
          <RobotAvatar size="xl" animated={true} />
        </div>

      </div>
  )
}*/
import { ArrowRight, Check } from "lucide-react"
import { StarBackground } from "./star-background"
import { RobotAvatar } from "./robot-avatar"

export function WelcomePage({ onGoToChat }) {
  const features = [
    "Talk to your computer in natural language",
    "Monitor system performance",
    "Analyze logs and errors",
    "Get smart suggestions and fixes",
    "Execute safe commands with permission",
  ]

  return (
    <div className="relative min-h-screen w-full overflow-x-hidden bg-[radial-gradient(circle_at_top,#171722_0%,#050508_45%,#020204_100%)] font-sans text-white">
      <div className="absolute inset-0">
        <StarBackground />
      </div>

      <div className="absolute inset-0 bg-[linear-gradient(135deg,rgba(255,255,255,0.04)_0%,transparent_30%,transparent_70%,rgba(255,255,255,0.03)_100%)]" />

      <div className="relative z-10 mx-auto flex h-screen w-full items-center px-4 py-6 sm:px-6 w-full">
        <div className="flex h-full w-full items-center gap-0 w-full">
          
          {/* LEFT CARD */}
          <div
            className=" 
              relative overflow-hidden rounded-[32px] border border-white/10
              bg-white/[0.06] p-4 shadow-[0_30px_80px_rgba(0,0,0,0.45)]
              backdrop-blur-xl sm:p-6 lg:p-8 
              lg:w-[66.67%] lg:max-h-[80vh]
            "
          >
            <div className="pointer-events-none absolute -top-24 -left-24 h-56 w-56 rounded-full bg-white/10 blur-3xl" />
            <div className="pointer-events-none absolute -bottom-24 -right-20 h-56 w-56 rounded-full bg-indigo-400/10 blur-3xl" />

            {/* Mobile top row: title + small avatar */}
            <div className="mb-4 flex items-start justify-between gap-4 lg:hidden">
              <img
                src="/titlewithstar.png"
                alt="Laptop Robot"
                className="h-auto w-[clamp(140px,48vw,220px)] object-contain"
              />

              <div className="shrink-0 pt-1">
                <RobotAvatar size="sm" animated={true} />
              </div>
            </div>

            {/* Desktop title */}
            <div className="mb-4 hidden lg:block">
              <img
                src="/titlewithstar.png"
                alt="Laptop Robot"
                className="h-auto w-[clamp(200px,18vw,300px)] object-contain"
              />
            </div>

            <h1 className="max-w-xl text-balance text-[clamp(1.4rem,2.2vw,2.2rem)] font-light leading-tight text-white">
              Welcome to Your{" "}
              <span className="font-semibold text-white">Laptop Robot</span>.
            </h1>

            <p className="mt-4 max-w-2xl text-[clamp(0.9rem,1.2vw,1rem)] leading-6 text-white/72">
              Turn your laptop into an intelligent assistant that can understand,
              monitor, and help you control your system.
            </p>

            <div className="mt-6">
              <h2 className="text-sm font-medium uppercase tracking-[0.22em] text-white/55">
                This app helps you
              </h2>

              <ul className="mt-3 space-y-2">
                {features.map((item, i) => (
                  <li
                    key={i}
                    className="flex items-start gap-3 text-sm leading-5 text-white/70 sm:text-[0.9rem]"
                  >
                    <span className="mt-1 inline-flex h-5 w-5 items-center justify-center rounded-full border border-white/10 bg-white/5">
                      <Check className="h-3.5 w-3.5 text-white/80" />
                    </span>
                    <span>{item}</span>
                  </li>
                ))}
              </ul>
            </div>

            <div className="mt-6 flex justify-start sm:justify-end lg:justify-end">
              <button
                type="button"
                onClick={onGoToChat}
                className="bg-white text-black px-6 py-2.5 rounded-full font-bold hover:bg-gray-200 transition-all active:scale-95 shadow-lg flex items-center gap-2"
              >
                go to chat
                <ArrowRight className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* RIGHT SIDE AVATAR — ONLY LARGE SCREENS */}
          <div className="relative hidden items-center justify-center lg:flex lg:w-[33.33%] pointer-events-none">
            <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
              <div className="h-[min(55vw,680px)] w-[min(55vw,680px)] rounded-full bg-white/5 blur-3xl" />
            </div>

            <div className="relative z-10 flex w-full justify-center pointer-events-none">
              <RobotAvatar size="xl" animated={true} />
            </div>
          </div>

        </div>
      </div>

      <img
        src="/icon.png"
        alt=""
        className="pointer-events-none absolute bottom-4 right-4 w-6 opacity-20"
      />
    </div>
  )
}
/*
import { cn } from "../lib/utils"

export function RobotAvatar({ className, size = "md", animated = true }) {
  const sizeMap = {
    sm: "clamp(64px, 11vw, 86px)",
    md: "clamp(110px, 14vw, 160px)",
    lg: "clamp(220px, 24vw, 320px)",
    xl: "clamp(320px, 42vw, 640px)",
  }

  const pixelSize = sizeMap[size] || sizeMap.md

  return (
    <div
      className={cn(
        "relative isolate flex-shrink-0 overflow-visible",
        animated && "robot-float",
        className
      )}
      style={{ width: pixelSize, height: pixelSize }}
    >
      <style>{`
        @keyframes robotFloat {
          0%, 100% { transform: translateY(0px); }
          50% { transform: translateY(-10px); }
        }
        .robot-float {
          animation: robotFloat 6s ease-in-out infinite;
        }
      `}</style>

      {/* soft glow behind robot * /}
      <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
        <div
          style={{
            width: "42%",
            height: "42%",
            background:
              "radial-gradient(circle at center, rgba(255,255,255,0.95) 0%, rgba(255,255,255,0.18) 45%, transparent 72%)",
            filter: "blur(22px)",
            borderRadius: "50%",
          }}
        />
      </div>

      {/* shadow under robot * /}
      <div
        className="pointer-events-none absolute bottom-0 left-1/2 -translate-x-1/2"
        style={{
          width: "52%",
          height: "12%",
          background: "rgba(0,0,0,0.28)",
          borderRadius: "50%",
          filter: "blur(12px)",
          zIndex: 0,
        }}
      />

      <img
        src="/robot.png"
        alt="Robot avatar"
        className="relative z-10 h-full w-full object-contain select-none"
        draggable="false"
      />
    </div>
  )
}*/
