// ReadingLevelSelector — DESIGN.md "Specialty UI Elements > Reading Level
// Selector": shadcn ToggleGroup with radius-md segments, accent-verified
// fill on the selected segment. PRD.md Core Capability 3 defines the two
// selectable levels (Simplified, ELI5); "Academic" (the original text) is
// always available as a baseline comparison via the separate
// TextComparisonToggle component, not as a third option here — see
// ResultPage.jsx for how the two toggles work together.
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
        // Radix fires onValueChange with "" when the user clicks the
        // already-selected item (would deselect it) — ignore that so the
        // selector always has exactly one active level, matching PRD.md's
        // "user selects target reading level" (a required choice, not
        // optional).
        if (next) onChange(next)
      }}
      className="inline-flex gap-1 rounded-md border border-border-default bg-surface-raised p-1"
      aria-label="Reading level"
    >
      {LEVELS.map((level) => (
        <ToggleGroupPrimitive.Item
          key={level.value}
          value={level.value}
          className={cn(
            "min-h-[44px] rounded-sm px-4 text-body font-medium text-ink-secondary transition-colors",
            "data-[state=on]:bg-accent-verified data-[state=on]:text-white",
          )}
        >
          {level.label}
        </ToggleGroupPrimitive.Item>
      ))}
    </ToggleGroupPrimitive.Root>
  )
}
