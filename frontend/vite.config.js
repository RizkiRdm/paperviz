import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'

// PaperViz frontend build config. Two things matter here beyond the
// defaults: the Tailwind v4 Vite plugin (styles are generated at build
// time, no separate PostCSS config needed), and a proxy to the Go backend
// during local dev so `npm run dev` can call /api/* without CORS headaches.
// In production the built assets are served directly by the Go binary
// (ARCHITECTURE.md: "single binary" architecture style) — no proxy needed
// there, since frontend and backend share an origin.
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
