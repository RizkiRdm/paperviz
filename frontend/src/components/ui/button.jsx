// Button — matches DESIGN.md "Components > Buttons" exactly:
//   Primary:   --accent-verified fill, white text, radius-md, shadow-card on hover only.
//   Secondary: transparent fill, --ink-primary text, 1px --border-default border, radius-md.
//   Disabled:  0.4 opacity, no hover transform.
//
// asChild lets a caller render this styling on a different element (e.g. an
// <a> for a "copy link" action) without wrapping in an extra <button> — the
// standard shadcn/ui pattern via Radix's Slot primitive.
import { Slot } from "@radix-ui/react-slot"
import { cn } from "@/lib/utils"

const variantClasses = {
  primary:
    "bg-accent-verified text-white hover:shadow-[0_1px_2px_rgba(20,23,31,0.04),0_4px_12px_rgba(20,23,31,0.06)]",
  secondary:
    "bg-transparent text-ink-primary border border-border-default hover:bg-surface-raised",
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
        // DESIGN.md "Do's and Don'ts": min 44x44px touch targets.
        "inline-flex min-h-[44px] items-center justify-center rounded-md px-4 text-body font-medium transition-shadow",
        "disabled:pointer-events-none disabled:opacity-40",
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  )
}
