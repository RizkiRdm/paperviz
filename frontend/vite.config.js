import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// PaperViz frontend build config. Two things matter here beyond the
// defaults: the Tailwind v4 Vite plugin (styles are generated at build
// time, no separate PostCSS config needed), and a proxy to the Go backend
// during local dev so `npm run dev` can call /api/* without CORS headaches.
// The Go binary serves the API only — the SPA is deployed separately (Cloudflare
// Pages today) and talks to it cross-origin via VITE_API_BASE_URL, with the
// backend allowlisting that origin in ALLOWED_ORIGINS. Local dev leaves
// VITE_API_BASE_URL unset so relative /api paths go through the proxy below.
export default defineConfig(({ mode }) => {
  // Repo root is the canonical local env location (.env.development lives
  // there per AGENTS.md); without envDir Vite only reads frontend/ and the
  // dev server boots without VITE_CLERK_PUBLISHABLE_KEY (guard screen).
  const repoRoot = path.resolve(__dirname, '..')
  const env = loadEnv(mode, repoRoot, '')
  const backendPort = env.PORT || "8080"

  return {
    envDir: repoRoot,
    plugins: [react(), tailwindcss()],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      proxy: {
        '/api': `http://localhost:${backendPort}`,
      },
    },
    build: {
      outDir: 'dist',
    },
  }
})
