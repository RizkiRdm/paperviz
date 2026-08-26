// Upgrade CTA — shown when user has reached their plan limit.
import { Button } from "./ui/button"

export default function UpgradeCta() {
  return (
    <div className="rounded-[12px] border border-[#e5e5e5] bg-white p-5 text-center">
      <p className="text-sm font-semibold text-[#0a0a0a] mb-1">
        You've reached your monthly limit
      </p>
      <p className="text-xs text-[#737373] mb-4">
        Upgrade to process more papers this month.
      </p>
      <div className="flex items-center justify-center gap-3">
        <Button variant="primary" asChild>
          <a href="#">Upgrade to Pro</a>
        </Button>
        <Button variant="secondary" asChild>
          <a href="#">Upgrade to Research</a>
        </Button>
      </div>
    </div>
  )
}
