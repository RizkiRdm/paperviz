// Radix Dialog wrapper styled per DESIGN.md (Elevated Feature Card: 16px radius,
// hairline border, subtle-2 ring shadow). Radix provides Esc dismiss, focus
// trap, and aria-modal semantics the hand-rolled overlay lacked.
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { X } from "lucide-react"
import { cn } from "@/lib/utils"

export const Dialog = DialogPrimitive.Root
export const DialogTrigger = DialogPrimitive.Trigger
export const DialogClose = DialogPrimitive.Close

export function DialogContent({ className, children, ...props }) {
  return (
    <DialogPrimitive.Portal>
      <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/20 backdrop-blur-[2px]" />
      <DialogPrimitive.Content
        className={cn(
          "fixed left-1/2 top-1/2 z-50 w-[calc(100vw-2rem)] max-w-md -translate-x-1/2 -translate-y-1/2",
          "rounded-[16px] border border-[#e5e5e5] bg-white p-6",
          "shadow-[rgba(0,0,0,0.1)_0px_0px_0px_4px] focus:outline-none",
          className,
        )}
        {...props}
      >
        {children}
        <DialogPrimitive.Close
          aria-label="Close"
          className="absolute right-4 top-4 rounded-full p-1 text-[#737373] hover:bg-[#f5f5f5] hover:text-[#0a0a0a] transition-colors cursor-pointer"
        >
          <X className="h-4 w-4" />
        </DialogPrimitive.Close>
      </DialogPrimitive.Content>
    </DialogPrimitive.Portal>
  )
}

export const DialogTitle = DialogPrimitive.Title
export const DialogDescription = DialogPrimitive.Description