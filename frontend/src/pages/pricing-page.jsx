// PricingPage — 3-column pricing tiers (Free / Pro / Research).
// Fires analytics view on mount, shows waitlist CTA for paid tiers.
import { useEffect } from "react"
import { Link } from "react-router-dom"
import { Button } from "@/components/ui/button"
import { Check } from "lucide-react"

const tiers = [
  {
    name: "Free",
    price: "$0",
    period: "forever",
    limit: "5 papers/month",
    features: [
      "Basic paper analysis",
      "Simplified text output",
      "Chapter-based charts",
      "7-day share links",
    ],
    cta: "Get Started",
    ctaLink: "/",
    primary: false,
  },
  {
    name: "Pro",
    price: "$29",
    period: "/month",
    limit: "50 papers/month",
    features: [
      "Everything in Free",
      "10x more papers",
      "Priority processing",
      "Advanced figure analysis",
    ],
    cta: "Upgrade to Pro",
    ctaLink: "#waitlist",
    primary: true,
  },
  {
    name: "Research",
    price: "Contact us",
    period: "",
    limit: "Custom volume",
    features: [
      "Everything in Pro",
      "200+ papers/month",
      "Custom support",
      "Priority processing",
    ],
    cta: "Contact Sales",
    ctaLink: "mailto:hello@paperviz.com",
    primary: false,
  },
]

export function PricingPage() {
  // Fire pricing page view event for conversion tracking
  useEffect(() => {
    fetch("/api/analytics/pricing-view", { method: "POST" }).catch(() => {})
  }, [])

  return (
    <div className="min-h-screen bg-white">
      {/* Header */}
      <header className="border-b border-[#e5e5e5]">
        <div className="mx-auto flex max-w-[1200px] items-center justify-between px-4 py-3">
          <Link to="/" className="font-satoshi text-lg font-medium text-[#0a0a0a]">
            PaperViz
          </Link>
          <div className="flex items-center gap-3">
            <Link
              to="/login"
              className="text-sm font-medium text-[#737373] hover:text-[#0a0a0a] transition-colors"
            >
              Log in
            </Link>
            <Button asChild variant="secondary" size="sm">
              <Link to="/signup">Sign up</Link>
            </Button>
          </div>
        </div>
      </header>

      {/* Main */}
      <main className="mx-auto max-w-[1200px] px-4 py-16">
        {/* Hero */}
        <div className="text-center mb-12">
          <h1 className="font-satoshi text-4xl sm:text-[48px] font-medium leading-[1.11] text-[#0a0a0a]">
            Simple pricing
          </h1>
          <p className="mt-3 text-base text-[#737373] max-w-md mx-auto">
            Start free. Upgrade when you need more papers per month.
          </p>
        </div>

        {/* Tiers grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {tiers.map((tier) => (
            <div
              key={tier.name}
              className={`flex flex-col rounded-[12px] border bg-white p-6 ${
                tier.primary
                  ? "border-[#2563eb] shadow-[rgba(37,99,235,0.08)_0px_0px_0px_1px]"
                  : "border-[#e5e5e5]"
              }`}
            >
              {/* Tier name + badge */}
              <div className="flex items-center gap-2 mb-4">
                <h2 className="font-satoshi text-xl font-medium text-[#0a0a0a]">
                  {tier.name}
                </h2>
                {tier.primary && (
                  <span className="inline-flex items-center rounded-full bg-[#dbeaff] px-2.5 py-0.5 text-[11px] font-medium text-[#2563eb]">
                    Popular
                  </span>
                )}
              </div>

              {/* Price */}
              <div className="mb-1 flex items-baseline gap-1">
                <span className="font-satoshi text-3xl font-medium text-[#0a0a0a]">
                  {tier.price}
                </span>
                {tier.period && (
                  <span className="text-sm text-[#737373]">{tier.period}</span>
                )}
              </div>
              <p className="text-xs text-[#737373] mb-6">{tier.limit}</p>

              {/* Features */}
              <ul className="flex flex-col gap-2.5 mb-8 flex-1">
                {tier.features.map((feature) => (
                  <li key={feature} className="flex items-start gap-2">
                    <Check className="mt-0.5 h-4 w-4 shrink-0 text-[#2563eb]" />
                    <span className="text-sm text-[#404040]">{feature}</span>
                  </li>
                ))}
              </ul>

              {/* CTA */}
              <Button
                asChild
                variant={tier.primary ? "primary" : "secondary"}
                className="w-full"
              >
                {tier.ctaLink.startsWith("/") ? (
                  <Link to={tier.ctaLink}>{tier.cta}</Link>
                ) : tier.ctaLink.startsWith("mailto:") ? (
                  <a href={tier.ctaLink}>{tier.cta}</a>
                ) : (
                  <a href={tier.ctaLink}>{tier.cta}</a>
                )}
              </Button>
            </div>
          ))}
        </div>

        {/* Bottom note */}
        <div className="mt-12 text-center">
          <p className="text-sm text-[#737373]">
            Payment integration coming soon. Join the waitlist to get early access.
          </p>
        </div>
      </main>
    </div>
  )
}
