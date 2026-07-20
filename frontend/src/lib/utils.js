// cn() merges Tailwind class strings intelligently — clsx handles
// conditional class logic (e.g. `cn("base", isActive && "active")`), and
// twMerge resolves conflicts when two classes touch the same CSS property
// (e.g. `cn("p-2", "p-4")` correctly keeps only "p-4" instead of emitting
// both and letting CSS source order decide). This is the standard
// shadcn/ui utility — every shadcn-style component in this project imports
// it instead of concatenating class strings by hand.
import { clsx } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs) {
  return twMerge(clsx(inputs))
}
