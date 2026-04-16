
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
    xl: "clamp(450px, 60vw, 340px)",
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
}