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

      stars = Array.from({ length: starCount }, () => {
        const isStar = Math.random() < 0.02
        return {
          x: Math.random() * canvas.width,
          y: Math.random() * canvas.height,
          size: isStar ? Math.random() * 10 + 8 : Math.random() * 2 + 0.5,
          opacity: isStar ? 1 : Math.random() * 0.3 + 0.2,
          speed: isStar ? Math.random() * 0.002 + 0.0005 : Math.random() * 0.02 + 0.005,
          phase: Math.random() * Math.PI * 2,
          drift: Math.random() * 0.01 - 0.005,
          type: isStar ? "star" : "dot",
        }
      })
    }

    function drawStarShape(ctx, x, y, outerRadius, opacity) {
      const spikes = 5
      const innerRadius = outerRadius * 0.38

      ctx.save()
      ctx.translate(x, y)
      ctx.rotate(-Math.PI / 2)

      ctx.beginPath()
      for (let i = 0; i < spikes * 2; i++) {
        const radius = i % 2 === 0 ? outerRadius : innerRadius
        const angle = (i * Math.PI) / spikes
        const px = Math.cos(angle) * radius
        const py = Math.sin(angle) * radius
        if (i === 0) ctx.moveTo(px, py)
        else ctx.lineTo(px, py)
      }
      ctx.closePath()

      ctx.fillStyle = `rgb(255,255,255)`
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

        // shape
        if (star.type === "star") {
          drawStarShape(ctx, star.x, star.y, star.size, opacity)
        } else {
          // small glow for dots only
          const gradient = ctx.createRadialGradient(
            star.x,
            star.y,
            0,
            star.x,
            star.y,
            star.size * 2
          )
          gradient.addColorStop(0, `rgba(255,255,255,${opacity * 0.3})`)
          gradient.addColorStop(1, "rgba(255,255,255,0)")
          ctx.fillStyle = gradient
          ctx.beginPath()
          ctx.arc(star.x, star.y, star.size * 2, 0, Math.PI * 2)
          ctx.fill()
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