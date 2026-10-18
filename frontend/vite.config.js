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
  // Load .env from project root so PORT in .env matches the backend proxy
  // target. Vite's config function runs before .envDir is processed, so we
  // need loadEnv explicitly here.
  const env = loadEnv(mode, path.resolve(__dirname, '..'), '')
  const backendPort = env.PORT || "8080"

  return {
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
