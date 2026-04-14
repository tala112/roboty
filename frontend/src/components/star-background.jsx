import { useEffect, useRef } from "react"

export function StarBackground() {
  const canvasRef = useRef(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext("2d")
    if (!ctx) return

    let stars = []
    let animationId

    const resizeCanvas = () => {
      canvas.width = window.innerWidth
      canvas.height = window.innerHeight

      const area = canvas.width * canvas.height
      const starCount = Math.floor(area / 8000)

      stars = Array.from({ length: starCount }, () => ({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        size: Math.random() * 2 + 0.5,
        opacity: Math.random(),
        speed: Math.random() * 0.02 + 0.005,
        phase: Math.random() * Math.PI * 2,
        drift: Math.random() * 0.1 - 0.05,
        type: Math.random() < 0.1 ? "star" : "dot",
      }))
    }

    function drawStarShape(ctx, x, y, size, opacity) {
      ctx.save()
      ctx.translate(x, y)

      const spikes = 4
      const outer = size
      const inner = size / 2

      let rot = (Math.PI / 2) * 3
      const step = Math.PI / spikes

      ctx.beginPath()
      ctx.moveTo(0, -outer)

      for (let i = 0; i < spikes; i++) {
        ctx.lineTo(Math.cos(rot) * outer, Math.sin(rot) * outer)
        rot += step

        ctx.lineTo(Math.cos(rot) * inner, Math.sin(rot) * inner)
        rot += step
      }

      ctx.closePath()
      ctx.fillStyle = `rgba(255,255,255,${opacity})`
      ctx.fill()

      ctx.restore()
    }

    const drawDot = (x, y, size, opacity) => {
      ctx.beginPath()
      ctx.arc(x, y, size, 0, Math.PI * 2)
      ctx.fillStyle = `rgba(255,255,255,${opacity})`
      ctx.fill()
    }

    const animate = () => {
      ctx.clearRect(0, 0, canvas.width, canvas.height)

      stars.forEach((star) => {
        // twinkle
        star.phase += star.speed
        const twinkle = Math.pow((Math.sin(star.phase) + 1) / 2, 2)
        const opacity = star.opacity * twinkle

        // drift
        star.y += star.drift

        // wrap around instead of reset random
        if (star.y < 0) star.y = canvas.height
        if (star.y > canvas.height) star.y = 0

        // glow
        const gradient = ctx.createRadialGradient(
          star.x,
          star.y,
          0,
          star.x,
          star.y,
          star.size * 3
        )

        gradient.addColorStop(0, `rgba(255,255,255,${opacity * 0.5})`)
        gradient.addColorStop(1, "rgba(255,255,255,0)")

        ctx.fillStyle = gradient
        ctx.beginPath()
        ctx.arc(star.x, star.y, star.size * 3, 0, Math.PI * 2)
        ctx.fill()

        // shape
        if (star.type === "star") {
          drawStarShape(ctx, star.x, star.y, star.size * 2, opacity)
        } else {
          drawDot(star.x, star.y, star.size, opacity)
        }
      })

      animationId = requestAnimationFrame(animate)
    }

    window.addEventListener("resize", resizeCanvas)
    resizeCanvas()
    animate()

    return () => {
      window.removeEventListener("resize", resizeCanvas)
      cancelAnimationFrame(animationId)
    }
  }, [])

  return (
    <canvas
      ref={canvasRef}
      className="fixed inset-0 pointer-events-none z-0"
      style={{ opacity: 0.7 }}
    />
  )
}