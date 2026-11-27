// ponytail: button variants per DESIGN.md (Filled Dark CTA #0a0a0a, Outlined #ffffff with #e5e5e5 border, Ghost)
import { Slot } from "@radix-ui/react-slot"
import { cn } from "@/lib/utils"

const variantClasses = {
  primary:
    "bg-[#0a0a0a] text-white hover:bg-[#171717] shadow-[rgba(0,0,0,0.05)_0px_1px_2px_0px]",
  secondary:
    "bg-white text-[#171717] border border-[#e5e5e5] hover:bg-[#f5f5f5] hover:border-[#d4d4d4]",
  outline:
    "bg-white text-[#171717] border border-[#e5e5e5] hover:bg-[#f5f5f5] hover:border-[#d4d4d4]",
  ghost:
    "bg-transparent text-[#171717] hover:bg-[#f5f5f5]",
  danger:
    "bg-red-50 text-red-600 border border-red-200 hover:bg-red-100",
}

const sizeClasses = {
  sm: "min-h-[32px] px-3 text-xs rounded-[6px] gap-1.5",
  md: "min-h-[40px] px-4 text-sm rounded-[8px] gap-2",
  lg: "min-h-[48px] px-5 text-base rounded-[8px] gap-2.5",
  icon: "h-[36px] w-[36px] p-0 rounded-[8px] justify-center",
}

export function Button({
  className,
  variant = "primary",
  size = "md",
  asChild = false,
  disabled = false,
  ...props
}) {
  const Comp = asChild ? Slot : "button"
  return (
    <Comp
      disabled={disabled}
      className={cn(
        "inline-flex items-center justify-center font-medium transition-all active:scale-[0.99] focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[#2563eb]",
        "disabled:pointer-events-none disabled:opacity-40",
        variantClasses[variant] || variantClasses.primary,
        sizeClasses[size] || sizeClasses.md,
        className,
      )}
      {...props}
    />
  )
}

