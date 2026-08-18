// Radix Tooltip wrapper styled per DESIGN.md (12px radius card, hairline
// border, paper-mist surface). Used for contextual help on badges and
// icon-adjacent controls.
import * as TooltipPrimitive from "@radix-ui/react-tooltip"
import { cn } from "@/lib/utils"

export const TooltipProvider = TooltipPrimitive.Provider
export const Tooltip = TooltipPrimitive.Root
export const TooltipTrigger = TooltipPrimitive.Trigger

export function TooltipContent({ className, sideOffset = 6, ...props }) {
  return (
    <TooltipPrimitive.Portal>
      <TooltipPrimitive.Content
        sideOffset={sideOffset}
        className={cn(
          "z-50 max-w-[240px] rounded-[8px] border border-[#e5e5e5] bg-white px-3 py-1.5",
          "text-[11px] font-medium leading-relaxed text-[#404040]",
          "shadow-[rgba(0,0,0,0.1)_0px_4px_6px_-1px,rgba(0,0,0,0.1)_0px_2px_4px_-2px]",
          "select-none",
          className,
        )}
        {...props}
      />
    </TooltipPrimitive.Portal>
  )
}