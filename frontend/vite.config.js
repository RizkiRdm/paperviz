import { defineConfig } from 'vite'
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
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
  build: {
    outDir: 'dist',
  },
})
