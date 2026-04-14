import { cn } from "../lib/utils"

export function RobotAvatar({ 
  className, 
  size = "md", 
  animated = true 
}) {
  const sizeMap = {
    sm: 120,
    md: 180,
    lg: 280,
    xl: 800,
  }

  const pixelSize = sizeMap[size]

  return (
    <div
      className={cn(
        "relative isolate", // ✅ important
        animated && "animate-float",
        className
      )}
      style={{ width: pixelSize, height: pixelSize }}
    >

      {/* 🔵 Glow (behind) */}
      <div
  className="absolute inset-0 flex items-center justify-center pointer-events-none"
  style={{ zIndex: 0 }}
>
  <div
    style={{
      width: "50%",
      height: "50%",
      background:
        "radial-gradient(ellipse at center, rgba(255,255,255,0.8) 0%, rgba(255,255,255,0.25) 35%, transparent 100%)",
      filter: "blur(30px)",
      transform: "scaleX(1) scaleY(1.3)", // 🔵 makes it oval
      borderRadius: "100%",
    }}
  />
</div>

      {/* 🟤 Shadow */}
      <div
        className="absolute bottom-0 left-1/2 -translate-x-1/2 pointer-events-none"
        style={{
          width: "45%",
          height: "18%",
          background: "rgba(0,0,0,0.3)",
          borderRadius: "50%",
          filter: "blur(12px)",
          zIndex: 0,
        }}
      />

      {/* 🤖 Image (on top) */}
      <img
        src="/robot.png"
        alt="Robot avatar"
        className="relative z-10 w-full h-full object-contain"
      />
    </div>
  )
}