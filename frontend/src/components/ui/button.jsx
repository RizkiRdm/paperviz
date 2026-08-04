// ponytail: button variants per DESIGN.md (Filled Dark CTA #0a0a0a, Outlined #ffffff with #e5e5e5 border, Ghost)
import { Slot } from "@radix-ui/react-slot"
import { cn } from "@/lib/utils"

const variantClasses = {
  primary:
    "bg-[#0a0a0a] text-white hover:bg-[#171717] shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]",
  secondary:
    "bg-white text-[#171717] border border-[#e5e5e5] hover:bg-[#f5f5f5]",
  ghost:
    "bg-transparent text-[#171717] hover:bg-[#f5f5f5]",
}

export function Button({
  className,
  variant = "primary",
  asChild = false,
  disabled = false,
  ...props
}) {
  const Comp = asChild ? Slot : "button"
  return (
    <Comp
      disabled={disabled}
      className={cn(
        "inline-flex min-h-[40px] items-center justify-center rounded-[8px] px-4 text-sm font-medium transition-all active:scale-[0.99]",
        "disabled:pointer-events-none disabled:opacity-40",
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  )
}

