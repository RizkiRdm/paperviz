// ponytail: pill selector using DESIGN.md 9999px radius and electric blue active state
import * as ToggleGroupPrimitive from "@radix-ui/react-toggle-group"
import { cn } from "@/lib/utils"

const LEVELS = [
  { value: "simplified", label: "Simplified" },
  { value: "eli5", label: "ELI5" },
]

export function ReadingLevelSelector({ value, onChange }) {
  return (
    <ToggleGroupPrimitive.Root
      type="single"
      value={value}
      onValueChange={(next) => {
        if (next) onChange(next)
      }}
      className="inline-flex gap-1 rounded-full border border-[#e5e5e5] bg-[#ffffff] p-1 shadow-xs"
      aria-label="Reading level"
    >
      {LEVELS.map((level) => (
        <ToggleGroupPrimitive.Item
          key={level.value}
          value={level.value}
          className={cn(
            "min-h-[36px] rounded-full px-4 text-xs font-medium text-[#737373] transition-colors hover:text-[#171717]",
            "data-[state=on]:bg-[#2563eb] data-[state=on]:text-white",
          )}
        >
          {level.label}
        </ToggleGroupPrimitive.Item>
      ))}
    </ToggleGroupPrimitive.Root>
  )
}

