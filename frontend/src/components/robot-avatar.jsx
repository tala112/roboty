
/*
import { cn } from "../lib/utils"

export function RobotAvatar({
  className,
  size = "md",
  animated = true,
}) {
  const sizeMap = {
    sm: "clamp(72px, 9vw, 110px)",
    md: "clamp(110px, 12vw, 160px)",
    lg: "clamp(250px, 18vw, 240px)",
    xl: "clamp(950px, 600vw, 540px)",
  }

  const pixelSize = sizeMap[size] || sizeMap.md

  return (
    <div
      className={cn(
        "relative isolate flex-shrink-0 overflow-visible",
        animated && "animate-float",
        className
      )}
      style={{ width: pixelSize, height: pixelSize }}
    >
      <div
        className="absolute inset-0 flex items-center justify-center pointer-events-none"
        style={{ zIndex: 0 }}
      >
        <div
          style={{
            width: "30%",
            height: "30%",
            background:
              "radial-gradient(ellipse at center, rgb(255, 255, 255) 10%, rgba(255,255,255,0.15) 55%, transparent 0%)",
            filter: "blur(20px)",
            borderRadius: "50%",
          }}
        />
      </div>

      <div
        className="absolute bottom-0 left-1/2 -translate-x-1/2 pointer-events-none"
        style={{
          width: "45%",
          height: "14%",
          background: "rgba(0,0,0,0.22)",
          borderRadius: "50%",
          filter: "blur(10px)",
          zIndex: 0,
        }}
      />

      <img
        src="/robot.png"
        alt="Robot avatar"
        className="relative z-10 h-full w-full object-contain"
      />
    </div>
  )
}*/
import { cn } from "../lib/utils"

export function RobotAvatar({ className, size = "md", animated = true }) {
  const sizeMap = {
    sm: "clamp(64px, 11vw, 86px)",
    md: "clamp(110px, 14vw, 160px)",
    lg: "clamp(220px, 24vw, 320px)",
    xl: "clamp(820px, 42vw, 640px)",
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

      {/* soft glow behind robot */}
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

      {/* shadow under robot */}
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
}