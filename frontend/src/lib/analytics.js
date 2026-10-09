// PostHog client. Loaded lazily so a blocked or slow PostHog CDN can never
// delay or break the app — analytics must not be on the render path.
//
// Events are fire-and-forget by design: PaperViz holds no third-party account,
// and an analytics failure must never surface to the user.
// A PostHog project key is a public, write-only ingest key — it is designed to
// ship in client JavaScript and grants no read access. Baking it in avoids
// threading a build argument through the Docker image for a non-secret.
const POSTHOG_KEY = "phc_nbnWEBTMCbCj4r9589UXegaXDhRDACqNdTDapbrVKkn4"
const POSTHOG_HOST = "https://us.i.posthog.com"

let clientPromise = null

// getClient resolves the PostHog SDK on first use and caches the promise.
export function getClient() {
  if (!POSTHOG_KEY) return Promise.resolve(null)
  if (clientPromise) return clientPromise

  clientPromise = import("posthog-js")
    .then(({ default: posthog }) => {
      posthog.init(POSTHOG_KEY, {
        api_host: POSTHOG_HOST,
        capture_pageview: true,
        capture_pageleave: true,
        persistence: "localStorage",
        person_profiles: "identified_only",
      })
      return posthog
    })
    .catch((err) => {
      console.warn("PostHog failed to load, analytics disabled", err)
      clientPromise = null
      return null
    })

  return clientPromise
}

// capture records an event. Errors are swallowed on purpose: telemetry must
// never break a user action, and a failed capture has no user-visible remedy.
export function capture(event, properties) {
  getClient().then((posthog) => {
    try {
      posthog?.capture(event, properties)
    } catch (err) {
      console.warn(`PostHog capture failed for ${event}`, err)
    }
  })
}

// identify ties later events to a known user.
export function identify(userId, properties) {
  getClient().then((posthog) => {
    try {
      posthog?.identify(userId, properties)
    } catch (err) {
      console.warn("PostHog identify failed", err)
    }
  })
}