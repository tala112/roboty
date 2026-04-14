import { useEffect, useState } from "react"

export function ThemeProvider({ children, defaultTheme = "dark", ...props }) {
  const [theme, setTheme] = useState(defaultTheme)

  useEffect(() => {
    const root = window.document.documentElement
    root.classList.remove("light", "dark")
    root.classList.add(theme)
  }, [theme])

  return (
    <div {...props}>
      {children}
    </div>
  )
}