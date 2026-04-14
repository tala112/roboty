import { forwardRef } from "react"
import { cn } from "../../lib/utils"

const ScrollArea = forwardRef(({ className, children, ...props }, ref) => (
  <div
    ref={ref}
    className={cn("relative", className)}
    {...props}
  >
    <div className="h-full w-full overflow-auto">
      {children}
    </div>
  </div>
))
ScrollArea.displayName = "ScrollArea"

export { ScrollArea }