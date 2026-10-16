// Radix Tabs wrapper styled per DESIGN.md (pill-adjacent tab bar: paper-mist
// track, ash border, white active tab, electric-blue active label).
import * as TabsPrimitive from "@radix-ui/react-tabs"
import { cn } from "@/lib/utils"

export const Tabs = TabsPrimitive.Root

export function TabsList({ className, ...props }) {
  return (
    <TabsPrimitive.List
      className={cn(
        "inline-flex items-center gap-1 rounded-[12px] border border-[#e5e5e5] bg-[#f5f5f5] p-1",
        className,
      )}
      {...props}
    />
  )
}

export function TabsTrigger({ className, ...props }) {
  return (
    <TabsPrimitive.Trigger
      className={cn(
        "rounded-[6px] px-3 py-1.5 text-sm font-medium text-[#737373] transition-colors cursor-pointer",
        "hover:text-[#171717] data-[state=active]:bg-white data-[state=active]:text-[#2563eb]",
        "focus-visible:outline-2 focus-visible:outline-[#2563eb] focus-visible:outline-offset-1",
        className,
      )}
      {...props}
    />
  )
}

export function TabsContent({ className, ...props }) {
  return (
    <TabsPrimitive.Content
      className={cn("mt-4 focus:outline-none", className)}
      {...props}
    />
  )
}
